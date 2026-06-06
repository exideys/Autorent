package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type ImageKind string

const (
	ImageKindCar  ImageKind = "car"
	ImageKindNews ImageKind = "news"
)

var (
	ErrNotConfigured      = errors.New("image storage is not configured")
	ErrInvalidImageKind   = errors.New("invalid image kind")
	ErrUnsupportedContent = errors.New("unsupported image content")
	ErrNotFound           = errors.New("image not found")
)

type GoogleDriveConfig struct {
	OAuthClientID     string
	OAuthClientSecret string
	OAuthRefreshToken string
	CarsFolderID      string
	NewsFolderID      string
	SupportFolderID   string
}

type GoogleDriveStorage struct {
	service         *drive.Service
	carsFolderID    string
	newsFolderID    string
	supportFolderID string
	authMode        string
}

type ImageUpload struct {
	Kind        ImageKind
	FileName    string
	ContentType string
	Body        io.Reader
}

type ImageUploadResult struct {
	FileID             string
	DriveWebViewURL    string
	DriveWebContentURL string
}

type FileUpload struct {
	FileName    string
	ContentType string
	Body        io.Reader
}

type FileUploadResult struct {
	FileID             string
	DriveWebViewURL    string
	DriveWebContentURL string
}

type DriveFile struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	Name          string
}

func NewGoogleDriveStorage(ctx context.Context, cfg GoogleDriveConfig) (*GoogleDriveStorage, error) {
	if strings.TrimSpace(cfg.CarsFolderID) == "" && strings.TrimSpace(cfg.NewsFolderID) == "" && strings.TrimSpace(cfg.SupportFolderID) == "" {
		return nil, nil
	}

	opts, authMode, err := googleDriveClientOptions(ctx, cfg)
	if err != nil {
		return nil, err
	}

	service, err := drive.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &GoogleDriveStorage{
		service:         service,
		carsFolderID:    strings.TrimSpace(cfg.CarsFolderID),
		newsFolderID:    strings.TrimSpace(cfg.NewsFolderID),
		supportFolderID: strings.TrimSpace(cfg.SupportFolderID),
		authMode:        authMode,
	}, nil
}

func (s *GoogleDriveStorage) AuthMode() string {
	if s == nil {
		return ""
	}
	return s.authMode
}

func googleDriveClientOptions(ctx context.Context, cfg GoogleDriveConfig) ([]option.ClientOption, string, error) {
	if hasOAuthConfig(cfg) {
		oauthConfig := &oauth2.Config{
			ClientID:     strings.TrimSpace(cfg.OAuthClientID),
			ClientSecret: strings.TrimSpace(cfg.OAuthClientSecret),
			Endpoint:     google.Endpoint,
			Scopes:       []string{drive.DriveFileScope},
		}
		tokenSource := oauthConfig.TokenSource(ctx, &oauth2.Token{
			RefreshToken: strings.TrimSpace(cfg.OAuthRefreshToken),
		})

		return []option.ClientOption{option.WithHTTPClient(oauth2.NewClient(ctx, tokenSource))}, "oauth", nil
	}
	if hasPartialOAuthConfig(cfg) {
		return nil, "", fmt.Errorf("%w: incomplete Google Drive OAuth configuration", ErrNotConfigured)
	}

	return nil, "", fmt.Errorf("%w: Google Drive OAuth is missing", ErrNotConfigured)
}

func hasOAuthConfig(cfg GoogleDriveConfig) bool {
	return strings.TrimSpace(cfg.OAuthClientID) != "" &&
		strings.TrimSpace(cfg.OAuthClientSecret) != "" &&
		strings.TrimSpace(cfg.OAuthRefreshToken) != ""
}

func hasPartialOAuthConfig(cfg GoogleDriveConfig) bool {
	return strings.TrimSpace(cfg.OAuthClientID) != "" ||
		strings.TrimSpace(cfg.OAuthClientSecret) != "" ||
		strings.TrimSpace(cfg.OAuthRefreshToken) != ""
}

func (s *GoogleDriveStorage) UploadImage(ctx context.Context, input ImageUpload) (*ImageUploadResult, error) {
	if s == nil || s.service == nil {
		return nil, ErrNotConfigured
	}
	if input.Body == nil {
		return nil, ErrUnsupportedContent
	}
	if !strings.HasPrefix(input.ContentType, "image/") {
		return nil, ErrUnsupportedContent
	}

	folderID, err := s.folderIDFor(input.Kind)
	if err != nil {
		return nil, err
	}

	fileName := uniqueImageName(input.FileName, input.ContentType)
	result, err := s.uploadFile(ctx, folderID, fileName, input.ContentType, input.Body)
	if err != nil {
		return nil, err
	}

	return &ImageUploadResult{
		FileID:             result.FileID,
		DriveWebViewURL:    result.DriveWebViewURL,
		DriveWebContentURL: result.DriveWebContentURL,
	}, nil
}

func (s *GoogleDriveStorage) UploadSupportAttachment(ctx context.Context, input FileUpload) (*FileUploadResult, error) {
	if s == nil || s.service == nil {
		return nil, ErrNotConfigured
	}
	if input.Body == nil || strings.TrimSpace(input.ContentType) == "" {
		return nil, ErrUnsupportedContent
	}
	if s.supportFolderID == "" {
		return nil, ErrNotConfigured
	}

	fileName := uniqueFileName(input.FileName)
	return s.uploadFile(ctx, s.supportFolderID, fileName, input.ContentType, input.Body)
}

func (s *GoogleDriveStorage) uploadFile(ctx context.Context, folderID string, fileName string, contentType string, body io.Reader) (*FileUploadResult, error) {
	file, err := s.service.Files.Create(&drive.File{
		Name:     fileName,
		MimeType: contentType,
		Parents:  []string{folderID},
	}).
		Media(body, googleapi.ContentType(contentType)).
		Fields("id, webViewLink, webContentLink").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return nil, mapGoogleAPIError(err)
	}

	return &FileUploadResult{
		FileID:             file.Id,
		DriveWebViewURL:    file.WebViewLink,
		DriveWebContentURL: file.WebContentLink,
	}, nil
}

func (s *GoogleDriveStorage) OpenDriveFile(ctx context.Context, fileID string) (*DriveFile, error) {
	if s == nil || s.service == nil {
		return nil, ErrNotConfigured
	}

	metadata, err := s.service.Files.Get(fileID).
		Fields("name, mimeType, size").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return nil, mapGoogleAPIError(err)
	}

	response, err := s.service.Files.Get(fileID).
		SupportsAllDrives(true).
		Context(ctx).
		Download()
	if err != nil {
		return nil, mapGoogleAPIError(err)
	}

	contentLength := int64(0)
	if metadata.Size > 0 {
		contentLength = metadata.Size
	} else if response.ContentLength > 0 {
		contentLength = response.ContentLength
	} else if sizeHeader := response.Header.Get("Content-Length"); sizeHeader != "" {
		if size, parseErr := strconv.ParseInt(sizeHeader, 10, 64); parseErr == nil {
			contentLength = size
		}
	}

	return &DriveFile{
		Body:          response.Body,
		ContentType:   metadata.MimeType,
		ContentLength: contentLength,
		Name:          metadata.Name,
	}, nil
}

func (s *GoogleDriveStorage) folderIDFor(kind ImageKind) (string, error) {
	switch kind {
	case ImageKindCar:
		if s.carsFolderID == "" {
			return "", ErrNotConfigured
		}
		return s.carsFolderID, nil
	case ImageKindNews:
		if s.newsFolderID == "" {
			return "", ErrNotConfigured
		}
		return s.newsFolderID, nil
	default:
		return "", ErrInvalidImageKind
	}
}

func uniqueImageName(originalName string, contentType string) string {
	extension := extensionForContentType(contentType)
	baseName := strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName))
	baseName = sanitizeFileName(baseName)
	if baseName == "" {
		baseName = "image"
	}

	return fmt.Sprintf(
		"%s-%s-%s%s",
		baseName,
		time.Now().UTC().Format("20060102-150405"),
		randomHex(4),
		extension,
	)
}

func uniqueFileName(originalName string) string {
	extension := sanitizeFileExtension(filepath.Ext(filepath.Base(originalName)))
	baseName := strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName))
	baseName = sanitizeFileName(baseName)
	if baseName == "" {
		baseName = "attachment"
	}

	return fmt.Sprintf(
		"%s-%s-%s%s",
		baseName,
		time.Now().UTC().Format("20060102-150405"),
		randomHex(4),
		extension,
	)
}

func sanitizeFileExtension(extension string) string {
	extension = strings.ToLower(strings.TrimSpace(extension))
	if len(extension) < 2 || len(extension) > 16 || !strings.HasPrefix(extension, ".") {
		return ""
	}

	for _, char := range extension[1:] {
		isAllowed := (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
		if !isAllowed {
			return ""
		}
	}

	return extension
}

func extensionForContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".img"
	}
}

func sanitizeFileName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	builder := strings.Builder{}
	lastWasDash := false

	for _, char := range value {
		isAllowed := (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
		if isAllowed {
			builder.WriteRune(char)
			lastWasDash = false
			continue
		}
		if !lastWasDash {
			builder.WriteByte('-')
			lastWasDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func randomHex(byteCount int) string {
	buffer := make([]byte, byteCount)
	if _, err := rand.Read(buffer); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buffer)
}

func mapGoogleAPIError(err error) error {
	var apiError *googleapi.Error
	if errors.As(err, &apiError) {
		switch apiError.Code {
		case 404:
			return ErrNotFound
		case 403:
			return fmt.Errorf("%w: google drive access denied", ErrNotConfigured)
		}
	}

	return err
}
