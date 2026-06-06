package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"
	"autorent-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

const (
	defaultMaxSupportAttachmentSize = 10 * 1024 * 1024
	maxSupportAttachments           = 5
	maxSupportMessageLength         = 4000
)

var (
	errSupportAttachmentTooLarge = errors.New("support attachment is too large")
	errSupportTooManyFiles       = errors.New("too many support attachments")
	errSupportInvalidPayload     = errors.New("invalid support message payload")
)

type SupportStore interface {
	GetOrCreateByUserID(ctx context.Context, userID int64) (*models.SupportConversation, error)
	GetByUserID(ctx context.Context, userID int64) (*models.SupportConversation, error)
	GetByID(ctx context.Context, id int64) (*models.SupportConversation, error)
	List(ctx context.Context) ([]models.SupportConversation, error)
	CreateMessage(
		ctx context.Context,
		conversationID int64,
		senderID int64,
		senderRole string,
		body string,
		attachments []models.SupportAttachmentInput,
	) (*models.SupportMessage, error)
	UpdateStatus(ctx context.Context, id int64, status string) (*models.SupportConversation, error)
	CanAccessAttachment(ctx context.Context, fileID string, userID int64, isAdmin bool) (bool, error)
}

type SupportFileStorage interface {
	UploadSupportAttachment(ctx context.Context, input storage.FileUpload) (*storage.FileUploadResult, error)
	OpenDriveFile(ctx context.Context, fileID string) (*storage.DriveFile, error)
}

type SupportHandler struct {
	store              SupportStore
	files              SupportFileStorage
	events             *SupportEventBroker
	maxAttachmentBytes int64
}

func RegisterSupportRoutes(
	router gin.IRouter,
	store SupportStore,
	files SupportFileStorage,
	events *SupportEventBroker,
	tokens *auth.TokenManager,
	maxAttachmentBytes int64,
) {
	handler := NewSupportHandler(store, files, events, maxAttachmentBytes)

	support := router.Group("/support")
	support.Use(RequireAuth(tokens))
	support.GET("/conversation", handler.GetMyConversation)
	support.POST("/messages", handler.CreateUserMessage)
	support.GET("/files/:file_id", handler.DownloadSupportFile)
	support.GET("/events", handler.StreamUserEvents)
}

func RegisterAdminSupportRoutes(
	router gin.IRouter,
	store SupportStore,
	files SupportFileStorage,
	events *SupportEventBroker,
	maxAttachmentBytes int64,
) {
	handler := NewSupportHandler(store, files, events, maxAttachmentBytes)

	support := router.Group("/support")
	support.GET("/conversations", handler.ListAdminConversations)
	support.GET("/conversations/:id", handler.GetAdminConversation)
	support.POST("/conversations/:id/messages", handler.CreateAdminMessage)
	support.PATCH("/conversations/:id/status", handler.UpdateAdminConversationStatus)
	support.GET("/events", handler.StreamAdminEvents)
}

func NewSupportHandler(store SupportStore, files SupportFileStorage, events *SupportEventBroker, maxAttachmentBytes int64) *SupportHandler {
	if maxAttachmentBytes <= 0 {
		maxAttachmentBytes = defaultMaxSupportAttachmentSize
	}

	return &SupportHandler{
		store:              store,
		files:              files,
		events:             events,
		maxAttachmentBytes: maxAttachmentBytes,
	}
}

func (h *SupportHandler) GetMyConversation(c *gin.Context) {
	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	conversation, err := h.store.GetByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"data": models.SupportConversation{
					UserID:   claims.UserID,
					Status:   models.SupportConversationStatusOpen,
					Messages: []models.SupportMessage{},
				},
			})
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load support conversation")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conversation})
}

func (h *SupportHandler) CreateUserMessage(c *gin.Context) {
	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	conversation, err := h.store.GetOrCreateByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to open support conversation")
		return
	}

	h.createMessage(c, conversation.ID, claims.UserID, models.SupportSenderUser, conversation.UserID)
}

func (h *SupportHandler) ListAdminConversations(c *gin.Context) {
	conversations, err := h.store.List(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load support conversations")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conversations})
}

func (h *SupportHandler) GetAdminConversation(c *gin.Context) {
	conversationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	conversation, err := h.store.GetByID(c.Request.Context(), conversationID)
	if err != nil {
		respondRepositoryError(c, err, "support conversation not found", "failed to load support conversation")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conversation})
}

func (h *SupportHandler) CreateAdminMessage(c *gin.Context) {
	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	conversationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	conversation, err := h.store.GetByID(c.Request.Context(), conversationID)
	if err != nil {
		respondRepositoryError(c, err, "support conversation not found", "failed to load support conversation")
		return
	}

	h.createMessage(c, conversationID, claims.UserID, models.SupportSenderAdmin, conversation.UserID)
}

func (h *SupportHandler) UpdateAdminConversationStatus(c *gin.Context) {
	conversationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid support conversation status")
		return
	}

	conversation, err := h.store.UpdateStatus(c.Request.Context(), conversationID, input.Status)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidInput):
			respondError(c, http.StatusBadRequest, "invalid support conversation status")
		default:
			respondRepositoryError(c, err, "support conversation not found", "failed to update support conversation")
		}
		return
	}
	if h.events != nil {
		h.events.Publish(SupportEvent{
			EventType:      supportEventTypeStatus,
			ConversationID: conversation.ID,
			UserID:         conversation.UserID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": conversation})
}

func (h *SupportHandler) StreamUserEvents(c *gin.Context) {
	if h.events == nil {
		respondError(c, http.StatusServiceUnavailable, "support realtime is not configured")
		return
	}

	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	events, unsubscribe := h.events.SubscribeUser(claims.UserID)
	defer unsubscribe()
	writeSupportEventStream(c.Writer, c.Request, events)
}

func (h *SupportHandler) StreamAdminEvents(c *gin.Context) {
	if h.events == nil {
		respondError(c, http.StatusServiceUnavailable, "support realtime is not configured")
		return
	}

	events, unsubscribe := h.events.SubscribeAdmin()
	defer unsubscribe()
	writeSupportEventStream(c.Writer, c.Request, events)
}

func (h *SupportHandler) DownloadSupportFile(c *gin.Context) {
	if h.files == nil {
		respondError(c, http.StatusServiceUnavailable, "support file storage is not configured")
		return
	}

	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	fileID := strings.TrimSpace(c.Param("file_id"))
	if !googleDriveFileIDPattern.MatchString(fileID) {
		respondError(c, http.StatusBadRequest, "invalid file id")
		return
	}

	canAccess, err := h.store.CanAccessAttachment(c.Request.Context(), fileID, claims.UserID, claims.Role == models.UserRoleAdmin)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load support file")
		return
	}
	if !canAccess {
		respondError(c, http.StatusNotFound, "support file not found")
		return
	}

	file, err := h.files.OpenDriveFile(c.Request.Context(), fileID)
	if err != nil {
		respondSupportStorageError(c, err)
		return
	}
	defer file.Body.Close()

	contentType := file.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, max-age=300")
	if file.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(file.ContentLength, 10))
	}
	if file.Name != "" {
		c.Header("Content-Disposition", `attachment; filename="`+sanitizeHeaderFileName(file.Name)+`"`)
	}

	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, file.Body)
}

func (h *SupportHandler) createMessage(c *gin.Context, conversationID int64, senderID int64, senderRole string, conversationUserID int64) {
	body, fileHeaders, err := h.parseMessageRequest(c)
	if err != nil {
		respondSupportMessageError(c, err)
		return
	}

	attachments, err := h.uploadSupportAttachments(c.Request.Context(), fileHeaders)
	if err != nil {
		respondSupportMessageError(c, err)
		return
	}

	message, err := h.store.CreateMessage(c.Request.Context(), conversationID, senderID, senderRole, body, attachments)
	if err != nil {
		respondRepositoryError(c, err, "support conversation not found", "failed to create support message")
		return
	}
	if h.events != nil {
		h.events.Publish(SupportEvent{
			ConversationID: conversationID,
			MessageID:      message.ID,
			UserID:         conversationUserID,
			SenderRole:     senderRole,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"data": message})
}

func (h *SupportHandler) parseMessageRequest(c *gin.Context) (string, []*multipart.FileHeader, error) {
	if c.ContentType() == "multipart/form-data" {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.multipartRequestMaxBytes())
		if err := c.Request.ParseMultipartForm(h.maxAttachmentBytes); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "too large") {
				return "", nil, errSupportAttachmentTooLarge
			}
			return "", nil, errSupportInvalidPayload
		}

		body := strings.TrimSpace(c.PostForm("message"))
		files := supportMultipartFiles(c.Request.MultipartForm)
		if err := validateSupportMessageBody(body, len(files)); err != nil {
			return "", nil, err
		}
		return body, files, nil
	}

	var input struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		return "", nil, errSupportInvalidPayload
	}

	body := strings.TrimSpace(input.Message)
	if err := validateSupportMessageBody(body, 0); err != nil {
		return "", nil, err
	}

	return body, nil, nil
}

func (h *SupportHandler) multipartRequestMaxBytes() int64 {
	return h.maxAttachmentBytes*maxSupportAttachments + maxSupportMessageLength + 1024*1024
}

func (h *SupportHandler) uploadSupportAttachments(ctx context.Context, fileHeaders []*multipart.FileHeader) ([]models.SupportAttachmentInput, error) {
	if len(fileHeaders) == 0 {
		return nil, nil
	}
	if len(fileHeaders) > maxSupportAttachments {
		return nil, errSupportTooManyFiles
	}
	if h.files == nil {
		return nil, storage.ErrNotConfigured
	}

	attachments := make([]models.SupportAttachmentInput, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		if fileHeader.Size <= 0 {
			return nil, storage.ErrUnsupportedContent
		}
		if fileHeader.Size > h.maxAttachmentBytes {
			return nil, errSupportAttachmentTooLarge
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, errSupportInvalidPayload
		}

		contentType, body, err := validatedSupportAttachmentBody(fileHeader, file)
		if err != nil {
			_ = file.Close()
			return nil, err
		}

		result, err := h.files.UploadSupportAttachment(ctx, storage.FileUpload{
			FileName:    fileHeader.Filename,
			ContentType: contentType,
			Body:        body,
		})
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			return nil, err
		}

		driveURL := strings.TrimSpace(result.DriveWebViewURL)
		var driveURLPtr *string
		if driveURL != "" {
			driveURLPtr = &driveURL
		}

		attachments = append(attachments, models.SupportAttachmentInput{
			FileName:    filepath.Base(fileHeader.Filename),
			ContentType: contentType,
			FileSize:    fileHeader.Size,
			DriveFileID: result.FileID,
			FileURL:     supportFileURL(result.FileID),
			DriveURL:    driveURLPtr,
		})
	}

	return attachments, nil
}

func validateSupportMessageBody(body string, fileCount int) error {
	if body == "" && fileCount == 0 {
		return errSupportInvalidPayload
	}
	if len([]rune(body)) > maxSupportMessageLength {
		return errSupportInvalidPayload
	}
	return nil
}

func supportMultipartFiles(form *multipart.Form) []*multipart.FileHeader {
	if form == nil {
		return nil
	}

	files := make([]*multipart.FileHeader, 0)
	for _, fieldName := range []string{"files", "file", "attachments"} {
		files = append(files, form.File[fieldName]...)
	}

	return files
}

func validatedSupportAttachmentBody(fileHeader *multipart.FileHeader, file multipart.File) (string, io.Reader, error) {
	header := make([]byte, 512)
	bytesRead, err := io.ReadFull(file, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", nil, err
	}
	if bytesRead == 0 {
		return "", nil, storage.ErrUnsupportedContent
	}

	header = header[:bytesRead]
	contentType := supportAttachmentContentType(
		http.DetectContentType(header),
		fileHeader.Header.Get("Content-Type"),
	)
	if !isAllowedSupportAttachmentContentType(contentType) {
		return "", nil, storage.ErrUnsupportedContent
	}

	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", nil, err
		}
		return contentType, file, nil
	}

	return contentType, io.MultiReader(bytes.NewReader(header), file), nil
}

func supportAttachmentContentType(detected string, declared string) string {
	detected = cleanContentType(detected)
	declared = cleanContentType(declared)

	if isAllowedSupportAttachmentContentType(declared) && (detected == "application/octet-stream" || detected == "application/zip") {
		return declared
	}
	if detected != "" {
		return detected
	}
	return declared
}

func cleanContentType(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if separator := strings.Index(contentType, ";"); separator >= 0 {
		contentType = strings.TrimSpace(contentType[:separator])
	}
	return contentType
}

func isAllowedSupportAttachmentContentType(contentType string) bool {
	switch cleanContentType(contentType) {
	case "image/jpeg",
		"image/png",
		"image/webp",
		"image/gif",
		"application/pdf",
		"text/plain",
		"text/csv",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return true
	default:
		return false
	}
}

func supportFileURL(fileID string) string {
	return "/api/support/files/" + fileID
}

func respondSupportMessageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errSupportInvalidPayload):
		respondError(c, http.StatusBadRequest, "invalid support message payload")
	case errors.Is(err, errSupportAttachmentTooLarge):
		respondError(c, http.StatusRequestEntityTooLarge, "support attachment is too large")
	case errors.Is(err, errSupportTooManyFiles):
		respondError(c, http.StatusBadRequest, "too many support attachments")
	case errors.Is(err, storage.ErrNotConfigured):
		respondError(c, http.StatusServiceUnavailable, "support file storage is not configured")
	case errors.Is(err, storage.ErrUnsupportedContent):
		respondError(c, http.StatusUnsupportedMediaType, "unsupported support file type")
	default:
		respondError(c, http.StatusInternalServerError, "failed to process support message")
	}
}

func respondSupportStorageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, storage.ErrNotConfigured):
		respondError(c, http.StatusServiceUnavailable, "support file storage is not configured")
	case errors.Is(err, storage.ErrNotFound):
		respondError(c, http.StatusNotFound, "support file not found")
	default:
		respondError(c, http.StatusInternalServerError, "failed to load support file")
	}
}
