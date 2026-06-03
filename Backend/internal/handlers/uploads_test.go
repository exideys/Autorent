package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"autorent-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

type fakeImageStorage struct {
	uploadFunc func(ctx context.Context, input storage.ImageUpload) (*storage.ImageUploadResult, error)
	openFunc   func(ctx context.Context, fileID string) (*storage.DriveFile, error)
}

func (f fakeImageStorage) UploadImage(ctx context.Context, input storage.ImageUpload) (*storage.ImageUploadResult, error) {
	return f.uploadFunc(ctx, input)
}

func (f fakeImageStorage) OpenDriveFile(ctx context.Context, fileID string) (*storage.DriveFile, error) {
	return f.openFunc(ctx, fileID)
}

func TestUploadCarImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedKind storage.ImageKind
	var capturedContentType string
	imageStorage := fakeImageStorage{
		uploadFunc: func(_ context.Context, input storage.ImageUpload) (*storage.ImageUploadResult, error) {
			capturedKind = input.Kind
			capturedContentType = input.ContentType
			body, err := io.ReadAll(input.Body)
			if err != nil {
				t.Fatalf("failed to read uploaded body: %v", err)
			}
			if len(body) == 0 || body[0] != 0x89 {
				t.Fatalf("unexpected uploaded body: %v", body)
			}

			return &storage.ImageUploadResult{
				FileID:          "drive-file-123",
				DriveWebViewURL: "https://drive.google.com/file/d/drive-file-123/view",
			}, nil
		},
	}

	router := gin.New()
	RegisterAdminUploadRoutes(router.Group("/api/admin"), imageStorage, 1024)

	req := multipartImageRequest(t, "/api/admin/uploads/car-image", "image", "car.png", testPNGBytes())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if capturedKind != storage.ImageKindCar || capturedContentType != "image/png" {
		t.Fatalf("unexpected upload input: kind=%s contentType=%s", capturedKind, capturedContentType)
	}
	if !strings.Contains(recorder.Body.String(), `"/api/images/google-drive/drive-file-123"`) {
		t.Fatalf("expected proxy image url in response, got %s", recorder.Body.String())
	}
}

func TestUploadImageRejectsUnsupportedContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imageStorage := fakeImageStorage{
		uploadFunc: func(context.Context, storage.ImageUpload) (*storage.ImageUploadResult, error) {
			t.Fatal("upload should not be called")
			return nil, nil
		},
	}

	router := gin.New()
	RegisterAdminUploadRoutes(router.Group("/api/admin"), imageStorage, 1024)

	req := multipartImageRequest(t, "/api/admin/uploads/news-image", "image", "notes.txt", []byte("not an image"))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, recorder.Code)
	}
}

func TestUploadImageMapsStorageErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "not configured", err: storage.ErrNotConfigured, wantStatus: http.StatusServiceUnavailable},
		{name: "not found", err: storage.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid kind", err: storage.ErrInvalidImageKind, wantStatus: http.StatusBadRequest},
		{name: "unsupported", err: storage.ErrUnsupportedContent, wantStatus: http.StatusUnsupportedMediaType},
		{name: "internal", err: errors.New("drive failed"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageStorage := fakeImageStorage{
				uploadFunc: func(context.Context, storage.ImageUpload) (*storage.ImageUploadResult, error) {
					return nil, tt.err
				},
			}

			router := gin.New()
			RegisterAdminUploadRoutes(router.Group("/api/admin"), imageStorage, 1024)

			req := multipartImageRequest(t, "/api/admin/uploads/news-image", "image", "news.png", testPNGBytes())
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestProxyGoogleDriveImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imageStorage := fakeImageStorage{
		openFunc: func(_ context.Context, fileID string) (*storage.DriveFile, error) {
			if fileID != "drive-file-123" {
				t.Fatalf("unexpected file id: %s", fileID)
			}

			return &storage.DriveFile{
				Body:          io.NopCloser(strings.NewReader("image-bytes")),
				ContentType:   "image/png",
				ContentLength: int64(len("image-bytes")),
				Name:          "car.png",
			}, nil
		},
	}

	router := gin.New()
	RegisterImageRoutes(router.Group("/api"), imageStorage)

	req, err := http.NewRequest(http.MethodGet, "/api/images/google-drive/drive-file-123", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("expected image/png content type, got %q", recorder.Header().Get("Content-Type"))
	}
	if recorder.Body.String() != "image-bytes" {
		t.Fatalf("unexpected body: %q", recorder.Body.String())
	}
}

func TestProxyGoogleDriveImageRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterImageRoutes(router.Group("/api"), fakeImageStorage{
		openFunc: func(context.Context, string) (*storage.DriveFile, error) {
			t.Fatal("open should not be called")
			return nil, nil
		},
	})

	req, err := http.NewRequest(http.MethodGet, "/api/images/google-drive/invalid!", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestProxyGoogleDriveImageMapsOpenError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterImageRoutes(router.Group("/api"), fakeImageStorage{
		openFunc: func(context.Context, string) (*storage.DriveFile, error) {
			return nil, storage.ErrNotFound
		},
	})

	req, err := http.NewRequest(http.MethodGet, "/api/images/google-drive/drive-file-123", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func multipartImageRequest(t *testing.T, target string, fieldName string, fileName string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, err := writer.CreateFormFile(fieldName, fileName)
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

func testPNGBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47,
		0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52,
	}
}
