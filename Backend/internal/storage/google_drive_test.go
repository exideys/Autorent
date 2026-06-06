package storage

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

func TestGoogleDriveClientOptionsRejectsPartialOAuthConfig(t *testing.T) {
	_, _, err := googleDriveClientOptions(context.Background(), GoogleDriveConfig{
		OAuthClientID: "client-id",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestGoogleDriveClientOptionsRejectsMissingAuthConfig(t *testing.T) {
	_, _, err := googleDriveClientOptions(context.Background(), GoogleDriveConfig{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestGoogleDriveClientOptionsUsesOAuthConfig(t *testing.T) {
	options, authMode, err := googleDriveClientOptions(context.Background(), GoogleDriveConfig{
		OAuthClientID:     "client-id",
		OAuthClientSecret: "client-secret",
		OAuthRefreshToken: "refresh-token",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(options) == 0 {
		t.Fatal("expected OAuth client option")
	}
	if authMode != "oauth" {
		t.Fatalf("expected oauth auth mode, got %q", authMode)
	}
}

func TestNewGoogleDriveStorageWithoutFoldersReturnsNil(t *testing.T) {
	storage, err := NewGoogleDriveStorage(context.Background(), GoogleDriveConfig{
		OAuthClientID:     "client-id",
		OAuthClientSecret: "client-secret",
		OAuthRefreshToken: "refresh-token",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if storage != nil {
		t.Fatalf("expected nil storage without folder ids, got %+v", storage)
	}
}

func TestNewGoogleDriveStorageWithFoldersRequiresOAuth(t *testing.T) {
	_, err := NewGoogleDriveStorage(context.Background(), GoogleDriveConfig{
		CarsFolderID: "cars-folder",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestGoogleDriveStorageAuthMode(t *testing.T) {
	var storage *GoogleDriveStorage
	if storage.AuthMode() != "" {
		t.Fatal("expected empty auth mode for nil storage")
	}

	storage = &GoogleDriveStorage{authMode: "oauth"}
	if storage.AuthMode() != "oauth" {
		t.Fatalf("expected oauth auth mode, got %q", storage.AuthMode())
	}
}

func TestGoogleDriveStorageFolderIDFor(t *testing.T) {
	storage := &GoogleDriveStorage{
		carsFolderID: "cars-folder",
		newsFolderID: "news-folder",
	}

	tests := []struct {
		name    string
		kind    ImageKind
		want    string
		wantErr error
	}{
		{name: "car", kind: ImageKindCar, want: "cars-folder"},
		{name: "news", kind: ImageKindNews, want: "news-folder"},
		{name: "invalid", kind: ImageKind("avatar"), wantErr: ErrInvalidImageKind},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folderID, err := storage.folderIDFor(tt.kind)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if folderID != tt.want {
				t.Fatalf("expected folder %q, got %q", tt.want, folderID)
			}
		})
	}

	_, err := (&GoogleDriveStorage{}).folderIDFor(ImageKindCar)
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for missing car folder, got %v", err)
	}
}

func TestGoogleDriveStorageUploadImageValidation(t *testing.T) {
	_, err := (*GoogleDriveStorage)(nil).UploadImage(context.Background(), ImageUpload{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for nil storage, got %v", err)
	}

	storage := &GoogleDriveStorage{}
	_, err = storage.UploadImage(context.Background(), ImageUpload{Body: strings.NewReader("image"), ContentType: "image/png"})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for missing service, got %v", err)
	}

	storage = &GoogleDriveStorage{service: &drive.Service{}, carsFolderID: "cars-folder"}
	_, err = storage.UploadImage(context.Background(), ImageUpload{ContentType: "image/png"})
	if !errors.Is(err, ErrUnsupportedContent) {
		t.Fatalf("expected ErrUnsupportedContent for nil body, got %v", err)
	}
	_, err = storage.UploadImage(context.Background(), ImageUpload{Body: strings.NewReader("text"), ContentType: "text/plain"})
	if !errors.Is(err, ErrUnsupportedContent) {
		t.Fatalf("expected ErrUnsupportedContent for non-image content, got %v", err)
	}
	_, err = storage.UploadImage(context.Background(), ImageUpload{
		Kind:        ImageKind("avatar"),
		Body:        strings.NewReader("image"),
		ContentType: "image/png",
	})
	if !errors.Is(err, ErrInvalidImageKind) {
		t.Fatalf("expected ErrInvalidImageKind, got %v", err)
	}

	storage = &GoogleDriveStorage{service: &drive.Service{}}
	_, err = storage.UploadImage(context.Background(), ImageUpload{
		Kind:        ImageKindCar,
		Body:        strings.NewReader("image"),
		ContentType: "image/png",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for missing folder, got %v", err)
	}
}

func TestGoogleDriveStorageUploadSupportAttachmentValidation(t *testing.T) {
	_, err := (*GoogleDriveStorage)(nil).UploadSupportAttachment(context.Background(), FileUpload{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for nil storage, got %v", err)
	}

	storage := &GoogleDriveStorage{}
	_, err = storage.UploadSupportAttachment(context.Background(), FileUpload{
		Body:        strings.NewReader("file"),
		ContentType: "application/pdf",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for missing service, got %v", err)
	}

	storage = &GoogleDriveStorage{service: &drive.Service{}, supportFolderID: "support-folder"}
	_, err = storage.UploadSupportAttachment(context.Background(), FileUpload{ContentType: "application/pdf"})
	if !errors.Is(err, ErrUnsupportedContent) {
		t.Fatalf("expected ErrUnsupportedContent for nil body, got %v", err)
	}
	_, err = storage.UploadSupportAttachment(context.Background(), FileUpload{
		Body:        strings.NewReader("file"),
		ContentType: "   ",
	})
	if !errors.Is(err, ErrUnsupportedContent) {
		t.Fatalf("expected ErrUnsupportedContent for empty content type, got %v", err)
	}

	storage = &GoogleDriveStorage{service: &drive.Service{}}
	_, err = storage.UploadSupportAttachment(context.Background(), FileUpload{
		Body:        strings.NewReader("file"),
		ContentType: "application/pdf",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for missing support folder, got %v", err)
	}
}

func TestGoogleDriveStorageOpenDriveFileRequiresService(t *testing.T) {
	_, err := (*GoogleDriveStorage)(nil).OpenDriveFile(context.Background(), "file-id")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for nil storage, got %v", err)
	}

	_, err = (&GoogleDriveStorage{}).OpenDriveFile(context.Background(), "file-id")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured for missing service, got %v", err)
	}
}

func TestImageNameHelpers(t *testing.T) {
	tests := []struct {
		contentType string
		extension   string
	}{
		{contentType: "image/jpeg", extension: ".jpg"},
		{contentType: " image/png ", extension: ".png"},
		{contentType: "image/webp", extension: ".webp"},
		{contentType: "image/gif", extension: ".gif"},
		{contentType: "application/octet-stream", extension: ".img"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			if got := extensionForContentType(tt.contentType); got != tt.extension {
				t.Fatalf("expected extension %q, got %q", tt.extension, got)
			}
		})
	}

	if got := sanitizeFileName(" BMW M5 / Competition! "); got != "bmw-m5-competition" {
		t.Fatalf("unexpected sanitized file name: %q", got)
	}
	if got := sanitizeFileName("%%%"); got != "" {
		t.Fatalf("expected empty sanitized file name, got %q", got)
	}
	if got := sanitizeFileExtension(".PDF"); got != ".pdf" {
		t.Fatalf("unexpected sanitized extension: %q", got)
	}
	if got := sanitizeFileExtension(".tar.gz"); got != "" {
		t.Fatalf("expected unsafe extension to be removed, got %q", got)
	}

	name := uniqueImageName("../BMW M5.png", "image/png")
	if !strings.HasPrefix(name, "bmw-m5-") || !strings.HasSuffix(name, ".png") || strings.Contains(name, "/") {
		t.Fatalf("unexpected unique image name: %q", name)
	}
	attachmentName := uniqueFileName("../Invoice Final.PDF")
	if !strings.HasPrefix(attachmentName, "invoice-final-") || !strings.HasSuffix(attachmentName, ".pdf") || strings.Contains(attachmentName, "/") {
		t.Fatalf("unexpected unique file name: %q", attachmentName)
	}
	attachmentName = uniqueFileName("%%%")
	if !strings.HasPrefix(attachmentName, "attachment-") || strings.Contains(attachmentName, ".") {
		t.Fatalf("unexpected fallback attachment name: %q", attachmentName)
	}

	if got := randomHex(4); len(got) != 8 {
		t.Fatalf("expected 8 hex chars, got %q", got)
	}
}

func TestMapGoogleAPIError(t *testing.T) {
	if err := mapGoogleAPIError(&googleapi.Error{Code: 404}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mapGoogleAPIError(&googleapi.Error{Code: 403}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}

	original := errors.New("network failed")
	if err := mapGoogleAPIError(original); !errors.Is(err, original) {
		t.Fatalf("expected original error, got %v", err)
	}
}
