package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"autorent-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSupportRepositoryGetOrCreateByUserIDReturnsExisting(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	lastMessageAt := now.Add(time.Minute)
	mock.ExpectQuery("SELECT").
		WithArgs(int64(42)).
		WillReturnRows(supportConversationRows().AddRow(
			7,
			42,
			models.SupportConversationStatusOpen,
			lastMessageAt,
			now,
			now,
			42,
			"Test",
			"User",
			"Test User",
			"user@example.com",
			4.8,
			2,
			models.UserRoleUser,
			"active",
			now,
			now,
		))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(7)).
		WillReturnRows(supportMessageRows())

	repo := NewSupportRepository(db)
	conversation, err := repo.GetOrCreateByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if conversation.ID != 7 || conversation.UserID != 42 || conversation.Status != models.SupportConversationStatusOpen {
		t.Fatalf("unexpected conversation: %+v", conversation)
	}
	if conversation.LastMessageAt == nil || !conversation.LastMessageAt.Equal(lastMessageAt) {
		t.Fatalf("unexpected last message time: %+v", conversation.LastMessageAt)
	}
	if conversation.User == nil || conversation.User.Email != "user@example.com" {
		t.Fatalf("unexpected conversation user: %+v", conversation.User)
	}
	if len(conversation.Messages) != 0 {
		t.Fatalf("expected empty messages, got %+v", conversation.Messages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSupportRepositoryGetOrCreateByUserIDCreatesMissingConversation(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery("SELECT").
		WithArgs(int64(42)).
		WillReturnRows(supportConversationRows())
	mock.ExpectExec("INSERT INTO support_conversations").
		WithArgs(int64(42), models.SupportConversationStatusOpen).
		WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(42)).
		WillReturnRows(supportConversationRows().AddRow(
			7,
			42,
			models.SupportConversationStatusOpen,
			nil,
			now,
			now,
			42,
			"Test",
			"User",
			"Test User",
			"user@example.com",
			5.0,
			0,
			models.UserRoleUser,
			"active",
			now,
			now,
		))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(7)).
		WillReturnRows(supportMessageRows())

	repo := NewSupportRepository(db)
	conversation, err := repo.GetOrCreateByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if conversation.ID != 7 || conversation.UserID != 42 {
		t.Fatalf("unexpected conversation: %+v", conversation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSupportRepositoryListLoadsMessagesAndAttachments(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	driveURL := "https://drive.google.com/file/d/support-file/view"
	mock.ExpectQuery("SELECT").
		WillReturnRows(supportConversationRows().AddRow(
			7,
			42,
			models.SupportConversationStatusOpen,
			now,
			now,
			now,
			42,
			"Test",
			"User",
			"Test User",
			"user@example.com",
			4.5,
			3,
			models.UserRoleUser,
			"active",
			now,
			now,
		))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(7)).
		WillReturnRows(supportMessageRows().AddRow(
			55,
			7,
			42,
			models.SupportSenderUser,
			"Need help",
			now,
		))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(55)).
		WillReturnRows(supportAttachmentRows().AddRow(
			70,
			55,
			"receipt.pdf",
			"application/pdf",
			int64(512),
			"support-file-123",
			"/api/support/files/support-file-123",
			driveURL,
			now,
		))

	repo := NewSupportRepository(db)
	conversations, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(conversations) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(conversations))
	}
	conversation := conversations[0]
	if len(conversation.Messages) != 1 {
		t.Fatalf("expected 1 message, got %+v", conversation.Messages)
	}
	message := conversation.Messages[0]
	if message.Body != "Need help" || message.SenderRole != models.SupportSenderUser {
		t.Fatalf("unexpected message: %+v", message)
	}
	if len(message.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %+v", message.Attachments)
	}
	attachment := message.Attachments[0]
	if attachment.DriveFileID != "support-file-123" || attachment.DriveURL == nil || *attachment.DriveURL != driveURL {
		t.Fatalf("unexpected attachment: %+v", attachment)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSupportRepositoryCreateMessageTrimsBodyAndStoresAttachments(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	driveURL := " https://drive.google.com/file/d/support-file/view "
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectExec("INSERT INTO support_messages").
		WithArgs(int64(7), int64(1), models.SupportSenderAdmin, "Hello").
		WillReturnResult(sqlmock.NewResult(55, 1))
	mock.ExpectExec("INSERT INTO support_attachments").
		WithArgs(
			int64(55),
			"receipt.pdf",
			"application/pdf",
			int64(512),
			"support-file-123",
			"/api/support/files/support-file-123",
			"https://drive.google.com/file/d/support-file/view",
		).
		WillReturnResult(sqlmock.NewResult(70, 1))
	mock.ExpectExec("UPDATE support_conversations").
		WithArgs(models.SupportSenderAdmin, models.SupportSenderUser, models.SupportConversationStatusOpen, sqlmock.AnyArg(), sqlmock.AnyArg(), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT").
		WithArgs(int64(55)).
		WillReturnRows(supportMessageRows().AddRow(
			55,
			7,
			1,
			models.SupportSenderAdmin,
			"Hello",
			now,
		))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(55)).
		WillReturnRows(supportAttachmentRows().AddRow(
			70,
			55,
			"receipt.pdf",
			"application/pdf",
			int64(512),
			"support-file-123",
			"/api/support/files/support-file-123",
			"https://drive.google.com/file/d/support-file/view",
			now,
		))

	repo := NewSupportRepository(db)
	message, err := repo.CreateMessage(context.Background(), 7, 1, " ADMIN ", " Hello ", []models.SupportAttachmentInput{
		{
			FileName:    " receipt.pdf ",
			ContentType: " application/pdf ",
			FileSize:    512,
			DriveFileID: " support-file-123 ",
			FileURL:     " /api/support/files/support-file-123 ",
			DriveURL:    &driveURL,
		},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if message.ID != 55 || message.Body != "Hello" || message.SenderRole != models.SupportSenderAdmin {
		t.Fatalf("unexpected message: %+v", message)
	}
	if len(message.Attachments) != 1 || message.Attachments[0].FileName != "receipt.pdf" {
		t.Fatalf("unexpected attachments: %+v", message.Attachments)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSupportRepositoryCreateMessageMissingConversationRollsBack(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(404)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))
	mock.ExpectRollback()

	repo := NewSupportRepository(db)
	_, err := repo.CreateMessage(context.Background(), 404, 42, models.SupportSenderUser, "Hello", nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSupportRepositoryUpdateStatusValidatesAndReloadsConversation(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectExec("UPDATE support_conversations").
		WithArgs(models.SupportConversationStatusClosed, int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(7)).
		WillReturnRows(supportConversationRows().AddRow(
			7,
			42,
			models.SupportConversationStatusClosed,
			nil,
			now,
			now,
			42,
			"Test",
			"User",
			"Test User",
			"user@example.com",
			5.0,
			0,
			models.UserRoleUser,
			"active",
			now,
			now,
		))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(7)).
		WillReturnRows(supportMessageRows())

	repo := NewSupportRepository(db)
	conversation, err := repo.UpdateStatus(context.Background(), 7, " CLOSED ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if conversation.Status != models.SupportConversationStatusClosed {
		t.Fatalf("unexpected status: %+v", conversation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSupportRepositoryUpdateStatusRejectsInvalidStatus(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSupportRepository(db)
	_, err := repo.UpdateStatus(context.Background(), 7, "pending")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSupportRepositoryCanAccessAttachment(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		isAdmin bool
		args    []driver.Value
		exists  int
		want    bool
	}{
		{
			name:    "admin checks any support attachment",
			userID:  1,
			isAdmin: true,
			args:    []driver.Value{"support-file-123"},
			exists:  1,
			want:    true,
		},
		{
			name:    "user is scoped to own conversation",
			userID:  42,
			isAdmin: false,
			args:    []driver.Value{"support-file-123", int64(42)},
			exists:  0,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := newMockDB(t)
			defer cleanup()

			mock.ExpectQuery("SELECT EXISTS").
				WithArgs(tt.args...).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(tt.exists))

			repo := NewSupportRepository(db)
			got, err := repo.CanAccessAttachment(context.Background(), " support-file-123 ", tt.userID, tt.isAdmin)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected access %v, got %v", tt.want, got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet sql expectations: %v", err)
			}
		})
	}
}

func supportConversationRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"status",
		"last_message_at",
		"created_at",
		"updated_at",
		"user_id",
		"first_name",
		"last_name",
		"name",
		"email",
		"rating",
		"rating_count",
		"role",
		"user_status",
		"user_created_at",
		"user_updated_at",
	})
}

func supportMessageRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"conversation_id",
		"sender_id",
		"sender_role",
		"body",
		"created_at",
	})
}

func supportAttachmentRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"message_id",
		"file_name",
		"content_type",
		"file_size",
		"drive_file_id",
		"file_url",
		"drive_url",
		"created_at",
	})
}
