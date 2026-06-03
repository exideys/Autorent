package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"autorent-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

const defaultMaxImageUploadSize = 10 * 1024 * 1024

var googleDriveFileIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,200}$`)

type ImageStorage interface {
	UploadImage(ctx context.Context, input storage.ImageUpload) (*storage.ImageUploadResult, error)
	OpenDriveFile(ctx context.Context, fileID string) (*storage.DriveFile, error)
}

type UploadHandler struct {
	images   ImageStorage
	maxBytes int64
}

func RegisterAdminUploadRoutes(router gin.IRouter, images ImageStorage, maxBytes int64) {
	handler := NewUploadHandler(images, maxBytes)

	uploads := router.Group("/uploads")
	uploads.POST("/car-image", handler.UploadCarImage)
	uploads.POST("/news-image", handler.UploadNewsImage)
}

func RegisterImageRoutes(router gin.IRouter, images ImageStorage) {
	handler := NewUploadHandler(images, defaultMaxImageUploadSize)

	router.GET("/images/google-drive/:file_id", handler.ProxyGoogleDriveImage)
}

func NewUploadHandler(images ImageStorage, maxBytes int64) *UploadHandler {
	if maxBytes <= 0 {
		maxBytes = defaultMaxImageUploadSize
	}

	return &UploadHandler{
		images:   images,
		maxBytes: maxBytes,
	}
}

func (h *UploadHandler) UploadCarImage(c *gin.Context) {
	h.uploadImage(c, storage.ImageKindCar)
}

func (h *UploadHandler) UploadNewsImage(c *gin.Context) {
	h.uploadImage(c, storage.ImageKindNews)
}

func (h *UploadHandler) ProxyGoogleDriveImage(c *gin.Context) {
	if h.images == nil {
		respondError(c, http.StatusServiceUnavailable, "image storage is not configured")
		return
	}

	fileID := strings.TrimSpace(c.Param("file_id"))
	if !googleDriveFileIDPattern.MatchString(fileID) {
		respondError(c, http.StatusBadRequest, "invalid image id")
		return
	}

	file, err := h.images.OpenDriveFile(c.Request.Context(), fileID)
	if err != nil {
		respondImageStorageError(c, err)
		return
	}
	defer file.Body.Close()

	contentType := file.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400")
	if file.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(file.ContentLength, 10))
	}
	if file.Name != "" {
		c.Header("Content-Disposition", `inline; filename="`+sanitizeHeaderFileName(file.Name)+`"`)
	}

	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, file.Body)
}

func (h *UploadHandler) uploadImage(c *gin.Context, kind storage.ImageKind) {
	if h.images == nil {
		respondError(c, http.StatusServiceUnavailable, "image storage is not configured")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBytes+1024)
	fileHeader, err := c.FormFile("image")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			respondError(c, http.StatusRequestEntityTooLarge, "image is too large")
			return
		}
		respondError(c, http.StatusBadRequest, "image file is required")
		return
	}
	if fileHeader.Size <= 0 {
		respondError(c, http.StatusBadRequest, "image file is empty")
		return
	}
	if fileHeader.Size > h.maxBytes {
		respondError(c, http.StatusRequestEntityTooLarge, "image is too large")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusBadRequest, "failed to read image file")
		return
	}
	defer file.Close()

	contentType, body, err := validatedImageBody(file)
	if err != nil {
		respondError(c, http.StatusUnsupportedMediaType, "unsupported image type")
		return
	}

	result, err := h.images.UploadImage(c.Request.Context(), storage.ImageUpload{
		Kind:        kind,
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		Body:        body,
	})
	if err != nil {
		respondImageStorageError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"image_url": googleDriveProxyImageURL(result.FileID),
			"file_id":   result.FileID,
			"drive_url": result.DriveWebViewURL,
		},
	})
}

func validatedImageBody(file multipart.File) (string, io.Reader, error) {
	header := make([]byte, 512)
	bytesRead, err := io.ReadFull(file, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", nil, err
	}
	if bytesRead == 0 {
		return "", nil, storage.ErrUnsupportedContent
	}

	header = header[:bytesRead]
	contentType := http.DetectContentType(header)
	if !isAllowedImageContentType(contentType) {
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

func isAllowedImageContentType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

func googleDriveProxyImageURL(fileID string) string {
	return "/api/images/google-drive/" + fileID
}

func respondImageStorageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, storage.ErrNotConfigured):
		respondError(c, http.StatusServiceUnavailable, "image storage is not configured")
	case errors.Is(err, storage.ErrNotFound):
		respondError(c, http.StatusNotFound, "image not found")
	case errors.Is(err, storage.ErrInvalidImageKind):
		respondError(c, http.StatusBadRequest, "invalid image type")
	case errors.Is(err, storage.ErrUnsupportedContent):
		respondError(c, http.StatusUnsupportedMediaType, "unsupported image type")
	default:
		respondError(c, http.StatusInternalServerError, "failed to process image")
	}
}

func sanitizeHeaderFileName(value string) string {
	value = strings.ReplaceAll(value, `"`, "")
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")
	if strings.TrimSpace(value) == "" {
		return "image"
	}
	return value
}
