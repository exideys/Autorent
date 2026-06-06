package handlers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"
	"autorent-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

type fakeSupportStore struct {
	getOrCreateByUserIDFunc func(ctx context.Context, userID int64) (*models.SupportConversation, error)
	getByUserIDFunc         func(ctx context.Context, userID int64) (*models.SupportConversation, error)
	getByIDFunc             func(ctx context.Context, id int64) (*models.SupportConversation, error)
	listFunc                func(ctx context.Context) ([]models.SupportConversation, error)
	createMessageFunc       func(
		ctx context.Context,
		conversationID int64,
		senderID int64,
		senderRole string,
		body string,
		attachments []models.SupportAttachmentInput,
	) (*models.SupportMessage, error)
	updateStatusFunc        func(ctx context.Context, id int64, status string) (*models.SupportConversation, error)
	canAccessAttachmentFunc func(ctx context.Context, fileID string, userID int64, isAdmin bool) (bool, error)
}

func (f *fakeSupportStore) GetOrCreateByUserID(ctx context.Context, userID int64) (*models.SupportConversation, error) {
	return f.getOrCreateByUserIDFunc(ctx, userID)
}

func (f *fakeSupportStore) GetByUserID(ctx context.Context, userID int64) (*models.SupportConversation, error) {
	return f.getByUserIDFunc(ctx, userID)
}

func (f *fakeSupportStore) GetByID(ctx context.Context, id int64) (*models.SupportConversation, error) {
	return f.getByIDFunc(ctx, id)
}

func (f *fakeSupportStore) List(ctx context.Context) ([]models.SupportConversation, error) {
	return f.listFunc(ctx)
}

func (f *fakeSupportStore) CreateMessage(
	ctx context.Context,
	conversationID int64,
	senderID int64,
	senderRole string,
	body string,
	attachments []models.SupportAttachmentInput,
) (*models.SupportMessage, error) {
	return f.createMessageFunc(ctx, conversationID, senderID, senderRole, body, attachments)
}

func (f *fakeSupportStore) UpdateStatus(ctx context.Context, id int64, status string) (*models.SupportConversation, error) {
	return f.updateStatusFunc(ctx, id, status)
}

func (f *fakeSupportStore) CanAccessAttachment(ctx context.Context, fileID string, userID int64, isAdmin bool) (bool, error) {
	return f.canAccessAttachmentFunc(ctx, fileID, userID, isAdmin)
}

type fakeSupportFileStorage struct {
	uploadFunc func(ctx context.Context, input storage.FileUpload) (*storage.FileUploadResult, error)
	openFunc   func(ctx context.Context, fileID string) (*storage.DriveFile, error)
}

func (f fakeSupportFileStorage) UploadSupportAttachment(ctx context.Context, input storage.FileUpload) (*storage.FileUploadResult, error) {
	return f.uploadFunc(ctx, input)
}

func (f fakeSupportFileStorage) OpenDriveFile(ctx context.Context, fileID string) (*storage.DriveFile, error) {
	return f.openFunc(ctx, fileID)
}

func TestCreateUserSupportMessageUsesAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 42, models.UserRoleUser)
	store := newFakeSupportStore()
	store.getOrCreateByUserIDFunc = func(_ context.Context, userID int64) (*models.SupportConversation, error) {
		if userID != 42 {
			t.Fatalf("expected user id 42, got %d", userID)
		}
		return &models.SupportConversation{ID: 7, UserID: userID}, nil
	}
	store.createMessageFunc = func(
		_ context.Context,
		conversationID int64,
		senderID int64,
		senderRole string,
		body string,
		attachments []models.SupportAttachmentInput,
	) (*models.SupportMessage, error) {
		if conversationID != 7 || senderID != 42 || senderRole != models.SupportSenderUser || body != "Need help" {
			t.Fatalf("unexpected message input: conversationID=%d senderID=%d senderRole=%s body=%q", conversationID, senderID, senderRole, body)
		}
		if len(attachments) != 0 {
			t.Fatalf("expected no attachments, got %+v", attachments)
		}
		return &models.SupportMessage{ID: 99, ConversationID: conversationID, SenderID: senderID, SenderRole: senderRole, Body: body}, nil
	}

	router := gin.New()
	RegisterSupportRoutes(router.Group("/api"), store, nil, nil, tokenManager, 1024)

	req, err := http.NewRequest(http.MethodPost, "/api/support/messages", strings.NewReader(`{"message":"Need help"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
}

func TestCreateUserSupportMessageUploadsAttachment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 42, models.UserRoleUser)
	store := newFakeSupportStore()
	store.getOrCreateByUserIDFunc = func(context.Context, int64) (*models.SupportConversation, error) {
		return &models.SupportConversation{ID: 7, UserID: 42}, nil
	}
	store.createMessageFunc = func(
		_ context.Context,
		_ int64,
		_ int64,
		_ string,
		_ string,
		attachments []models.SupportAttachmentInput,
	) (*models.SupportMessage, error) {
		if len(attachments) != 1 {
			t.Fatalf("expected one attachment, got %+v", attachments)
		}
		if attachments[0].DriveFileID != "support-file-123" || attachments[0].FileURL != "/api/support/files/support-file-123" {
			t.Fatalf("unexpected attachment metadata: %+v", attachments[0])
		}
		return &models.SupportMessage{ID: 99, ConversationID: 7, SenderID: 42, SenderRole: models.SupportSenderUser}, nil
	}
	fileStorage := fakeSupportFileStorage{
		uploadFunc: func(_ context.Context, input storage.FileUpload) (*storage.FileUploadResult, error) {
			if input.FileName != "support.txt" || input.ContentType != "text/plain" {
				t.Fatalf("unexpected upload input: fileName=%q contentType=%q", input.FileName, input.ContentType)
			}
			body, err := io.ReadAll(input.Body)
			if err != nil {
				t.Fatalf("failed to read upload body: %v", err)
			}
			if string(body) != "hello support" {
				t.Fatalf("unexpected upload body: %q", string(body))
			}
			return &storage.FileUploadResult{
				FileID:          "support-file-123",
				DriveWebViewURL: "https://drive.google.com/file/d/support-file-123/view",
			}, nil
		},
	}

	router := gin.New()
	RegisterSupportRoutes(router.Group("/api"), store, fileStorage, nil, tokenManager, 1024)

	req := multipartSupportRequest(t, "/api/support/messages", "Need help", "support.txt", []byte("hello support"))
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
}

func TestGetMySupportConversationNotFoundReturnsEmptyConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 42, models.UserRoleUser)
	store := newFakeSupportStore()
	store.getByUserIDFunc = func(_ context.Context, userID int64) (*models.SupportConversation, error) {
		if userID != 42 {
			t.Fatalf("expected user id 42, got %d", userID)
		}
		return nil, repository.ErrNotFound
	}

	router := gin.New()
	RegisterSupportRoutes(router.Group("/api"), store, nil, nil, tokenManager, 1024)

	req, err := http.NewRequest(http.MethodGet, "/api/support/conversation", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"user_id":42`) || !strings.Contains(recorder.Body.String(), `"status":"open"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestCreateUserSupportMessageRejectsEmptyPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 42, models.UserRoleUser)
	store := newFakeSupportStore()
	store.getOrCreateByUserIDFunc = func(context.Context, int64) (*models.SupportConversation, error) {
		return &models.SupportConversation{ID: 7, UserID: 42}, nil
	}
	store.createMessageFunc = func(context.Context, int64, int64, string, string, []models.SupportAttachmentInput) (*models.SupportMessage, error) {
		t.Fatal("message should not be created for an empty payload")
		return nil, nil
	}

	router := gin.New()
	RegisterSupportRoutes(router.Group("/api"), store, nil, nil, tokenManager, 1024)

	req, err := http.NewRequest(http.MethodPost, "/api/support/messages", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func TestDownloadSupportFileChecksConversationAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 42, models.UserRoleUser)
	store := newFakeSupportStore()
	store.canAccessAttachmentFunc = func(_ context.Context, fileID string, userID int64, isAdmin bool) (bool, error) {
		if fileID != "support-file-123" || userID != 42 || isAdmin {
			t.Fatalf("unexpected access check: fileID=%s userID=%d isAdmin=%v", fileID, userID, isAdmin)
		}
		return false, nil
	}
	fileStorage := fakeSupportFileStorage{
		openFunc: func(context.Context, string) (*storage.DriveFile, error) {
			t.Fatal("file should not be opened without access")
			return nil, nil
		},
	}

	router := gin.New()
	RegisterSupportRoutes(router.Group("/api"), store, fileStorage, nil, tokenManager, 1024)

	req, err := http.NewRequest(http.MethodGet, "/api/support/files/support-file-123", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestAdminCanDownloadSupportFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 1, models.UserRoleAdmin)
	store := newFakeSupportStore()
	store.canAccessAttachmentFunc = func(_ context.Context, fileID string, userID int64, isAdmin bool) (bool, error) {
		if fileID != "support-file-123" || userID != 1 || !isAdmin {
			t.Fatalf("unexpected access check: fileID=%s userID=%d isAdmin=%v", fileID, userID, isAdmin)
		}
		return true, nil
	}
	fileStorage := fakeSupportFileStorage{
		openFunc: func(_ context.Context, fileID string) (*storage.DriveFile, error) {
			if fileID != "support-file-123" {
				t.Fatalf("unexpected file id: %s", fileID)
			}
			return &storage.DriveFile{
				Body:          io.NopCloser(strings.NewReader("support-file-body")),
				ContentType:   "text/plain",
				ContentLength: int64(len("support-file-body")),
				Name:          "support.txt",
			}, nil
		},
	}

	router := gin.New()
	RegisterSupportRoutes(router.Group("/api"), store, fileStorage, nil, tokenManager, 1024)

	req, err := http.NewRequest(http.MethodGet, "/api/support/files/support-file-123", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != "support-file-body" {
		t.Fatalf("unexpected body: %q", recorder.Body.String())
	}
}

func TestAdminListSupportConversations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeSupportStore()
	store.listFunc = func(context.Context) ([]models.SupportConversation, error) {
		return []models.SupportConversation{
			{ID: 7, UserID: 42, Status: models.SupportConversationStatusOpen},
		}, nil
	}

	router := gin.New()
	RegisterAdminSupportRoutes(router.Group("/api/admin"), store, nil, nil, 1024)

	req, err := http.NewRequest(http.MethodGet, "/api/admin/support/conversations", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"id":7`) || !strings.Contains(recorder.Body.String(), `"status":"open"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestAdminCreateSupportMessageUsesAuthenticatedAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 1, models.UserRoleAdmin)
	store := newFakeSupportStore()
	store.getByIDFunc = func(_ context.Context, id int64) (*models.SupportConversation, error) {
		if id != 7 {
			t.Fatalf("expected conversation id 7, got %d", id)
		}
		return &models.SupportConversation{ID: 7, UserID: 42, Status: models.SupportConversationStatusOpen}, nil
	}
	store.createMessageFunc = func(
		_ context.Context,
		conversationID int64,
		senderID int64,
		senderRole string,
		body string,
		attachments []models.SupportAttachmentInput,
	) (*models.SupportMessage, error) {
		if conversationID != 7 || senderID != 1 || senderRole != models.SupportSenderAdmin || body != "We are checking" {
			t.Fatalf("unexpected message input: conversationID=%d senderID=%d senderRole=%s body=%q", conversationID, senderID, senderRole, body)
		}
		if len(attachments) != 0 {
			t.Fatalf("expected no attachments, got %+v", attachments)
		}
		return &models.SupportMessage{ID: 99, ConversationID: conversationID, SenderID: senderID, SenderRole: senderRole, Body: body}, nil
	}

	router := gin.New()
	admin := router.Group("/api/admin")
	admin.Use(RequireAdmin(tokenManager))
	RegisterAdminSupportRoutes(admin, store, nil, nil, 1024)

	req, err := http.NewRequest(http.MethodPost, "/api/admin/support/conversations/7/messages", strings.NewReader(`{"message":"We are checking"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
}

func TestAdminUpdateSupportConversationStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeSupportStore()
	store.updateStatusFunc = func(_ context.Context, id int64, status string) (*models.SupportConversation, error) {
		if id != 7 || status != models.SupportConversationStatusClosed {
			t.Fatalf("unexpected status update: id=%d status=%q", id, status)
		}
		return &models.SupportConversation{
			ID:     id,
			UserID: 42,
			Status: models.SupportConversationStatusClosed,
		}, nil
	}

	router := gin.New()
	RegisterAdminSupportRoutes(router.Group("/api/admin"), store, nil, NewSupportEventBroker(), 1024)

	req, err := http.NewRequest(http.MethodPatch, "/api/admin/support/conversations/7/status", strings.NewReader(`{"status":"closed"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func newFakeSupportStore() *fakeSupportStore {
	return &fakeSupportStore{
		getOrCreateByUserIDFunc: func(context.Context, int64) (*models.SupportConversation, error) {
			return &models.SupportConversation{}, nil
		},
		getByUserIDFunc: func(context.Context, int64) (*models.SupportConversation, error) {
			return &models.SupportConversation{}, nil
		},
		getByIDFunc: func(context.Context, int64) (*models.SupportConversation, error) {
			return &models.SupportConversation{}, nil
		},
		listFunc: func(context.Context) ([]models.SupportConversation, error) {
			return []models.SupportConversation{}, nil
		},
		createMessageFunc: func(context.Context, int64, int64, string, string, []models.SupportAttachmentInput) (*models.SupportMessage, error) {
			return &models.SupportMessage{}, nil
		},
		updateStatusFunc: func(context.Context, int64, string) (*models.SupportConversation, error) {
			return &models.SupportConversation{}, nil
		},
		canAccessAttachmentFunc: func(context.Context, string, int64, bool) (bool, error) {
			return true, nil
		},
	}
}

func multipartSupportRequest(t *testing.T, target string, message string, fileName string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("message", message); err != nil {
		t.Fatalf("failed to write message field: %v", err)
	}
	fileWriter, err := writer.CreateFormFile("files", fileName)
	if err != nil {
		t.Fatalf("failed to create multipart file: %v", err)
	}
	if _, err := fileWriter.Write(content); err != nil {
		t.Fatalf("failed to write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, target, &body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
