package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"autorent-backend/internal/models"
)

type SupportRepository struct {
	db *sql.DB
}

var ErrInvalidInput = errors.New("invalid input")

func NewSupportRepository(db *sql.DB) *SupportRepository {
	return &SupportRepository{db: db}
}

func (r *SupportRepository) GetOrCreateByUserID(ctx context.Context, userID int64) (*models.SupportConversation, error) {
	conversation, err := r.GetByUserID(ctx, userID)
	if err == nil {
		return conversation, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO support_conversations (user_id, status)
		VALUES (?, ?)
	`, userID, models.SupportConversationStatusOpen)
	if err != nil {
		if isDuplicateEntry(err) {
			return r.GetByUserID(ctx, userID)
		}
		return nil, err
	}

	return r.GetByUserID(ctx, userID)
}

func (r *SupportRepository) GetByUserID(ctx context.Context, userID int64) (*models.SupportConversation, error) {
	rows, err := r.db.QueryContext(ctx, baseSupportConversationQuery()+" WHERE sc.user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations, err := scanSupportConversations(rows)
	if err != nil {
		return nil, err
	}
	if len(conversations) == 0 {
		return nil, ErrNotFound
	}

	if err := r.loadConversationMessages(ctx, &conversations[0]); err != nil {
		return nil, err
	}

	return &conversations[0], nil
}

func (r *SupportRepository) GetByID(ctx context.Context, id int64) (*models.SupportConversation, error) {
	rows, err := r.db.QueryContext(ctx, baseSupportConversationQuery()+" WHERE sc.id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations, err := scanSupportConversations(rows)
	if err != nil {
		return nil, err
	}
	if len(conversations) == 0 {
		return nil, ErrNotFound
	}

	if err := r.loadConversationMessages(ctx, &conversations[0]); err != nil {
		return nil, err
	}

	return &conversations[0], nil
}

func (r *SupportRepository) List(ctx context.Context) ([]models.SupportConversation, error) {
	rows, err := r.db.QueryContext(ctx, baseSupportConversationQuery()+`
		ORDER BY COALESCE(sc.last_message_at, sc.created_at) DESC, sc.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations, err := scanSupportConversations(rows)
	if err != nil {
		return nil, err
	}

	for index := range conversations {
		if err := r.loadConversationMessages(ctx, &conversations[index]); err != nil {
			return nil, err
		}
	}

	return conversations, nil
}

func (r *SupportRepository) CreateMessage(
	ctx context.Context,
	conversationID int64,
	senderID int64,
	senderRole string,
	body string,
	attachments []models.SupportAttachmentInput,
) (*models.SupportMessage, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	var exists int
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM support_conversations WHERE id = ?)", conversationID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists != 1 {
		return nil, ErrNotFound
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO support_messages (conversation_id, sender_id, sender_role, body)
		VALUES (?, ?, ?, ?)
	`, conversationID, senderID, supportSenderRoleOrDefault(senderRole), strings.TrimSpace(body))
	if err != nil {
		return nil, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, attachment := range attachments {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO support_attachments (
				message_id, file_name, content_type, file_size, drive_file_id, file_url, drive_url
			)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
			messageID,
			strings.TrimSpace(attachment.FileName),
			strings.TrimSpace(attachment.ContentType),
			attachment.FileSize,
			strings.TrimSpace(attachment.DriveFileID),
			strings.TrimSpace(attachment.FileURL),
			trimmedStringPtrValue(attachment.DriveURL),
		); err != nil {
			return nil, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE support_conversations
		SET
			status = CASE WHEN ? = ? THEN ? ELSE status END,
			last_message_at = ?,
			updated_at = ?
		WHERE id = ?
	`,
		supportSenderRoleOrDefault(senderRole),
		models.SupportSenderUser,
		models.SupportConversationStatusOpen,
		time.Now().UTC(),
		time.Now().UTC(),
		conversationID,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.getMessageByID(ctx, messageID)
}

func (r *SupportRepository) UpdateStatus(ctx context.Context, id int64, status string) (*models.SupportConversation, error) {
	normalizedStatus, ok := normalizeSupportConversationStatus(status)
	if !ok {
		return nil, ErrInvalidInput
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE support_conversations
		SET status = ?
		WHERE id = ?
	`, normalizedStatus, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *SupportRepository) CanAccessAttachment(ctx context.Context, fileID string, userID int64, isAdmin bool) (bool, error) {
	var exists int
	var err error

	if isAdmin {
		err = r.db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM support_attachments
				WHERE drive_file_id = ?
			)
		`, strings.TrimSpace(fileID)).Scan(&exists)
	} else {
		err = r.db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM support_attachments sa
				JOIN support_messages sm ON sm.id = sa.message_id
				JOIN support_conversations sc ON sc.id = sm.conversation_id
				WHERE sa.drive_file_id = ? AND sc.user_id = ?
			)
		`, strings.TrimSpace(fileID), userID).Scan(&exists)
	}
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (r *SupportRepository) getMessageByID(ctx context.Context, id int64) (*models.SupportMessage, error) {
	rows, err := r.db.QueryContext(ctx, baseSupportMessageQuery()+" WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages, err := scanSupportMessages(rows)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, ErrNotFound
	}

	attachments, err := r.loadAttachments(ctx, messages[0].ID)
	if err != nil {
		return nil, err
	}
	messages[0].Attachments = attachments

	return &messages[0], nil
}

func (r *SupportRepository) loadConversationMessages(ctx context.Context, conversation *models.SupportConversation) error {
	rows, err := r.db.QueryContext(ctx, baseSupportMessageQuery()+`
		WHERE conversation_id = ?
		ORDER BY created_at ASC, id ASC
	`, conversation.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	messages, err := scanSupportMessages(rows)
	if err != nil {
		return err
	}

	for index := range messages {
		attachments, err := r.loadAttachments(ctx, messages[index].ID)
		if err != nil {
			return err
		}
		messages[index].Attachments = attachments
	}

	conversation.Messages = messages
	return nil
}

func (r *SupportRepository) loadAttachments(ctx context.Context, messageID int64) ([]models.SupportAttachment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			message_id,
			file_name,
			content_type,
			file_size,
			drive_file_id,
			file_url,
			drive_url,
			created_at
		FROM support_attachments
		WHERE message_id = ?
		ORDER BY id ASC
	`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]models.SupportAttachment, 0)
	for rows.Next() {
		var attachment models.SupportAttachment
		var driveURL sql.NullString

		if err := rows.Scan(
			&attachment.ID,
			&attachment.MessageID,
			&attachment.FileName,
			&attachment.ContentType,
			&attachment.FileSize,
			&attachment.DriveFileID,
			&attachment.FileURL,
			&driveURL,
			&attachment.CreatedAt,
		); err != nil {
			return nil, err
		}

		if driveURL.Valid {
			attachment.DriveURL = &driveURL.String
		}
		attachments = append(attachments, attachment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

func baseSupportConversationQuery() string {
	return `
		SELECT
			sc.id,
			sc.user_id,
			sc.status,
			sc.last_message_at,
			sc.created_at,
			sc.updated_at,
			u.id,
			u.first_name,
			u.last_name,
			TRIM(CONCAT_WS(' ', u.first_name, u.last_name)) AS name,
			u.email,
			u.rating,
			u.rating_count,
			u.role,
			u.status,
			u.created_at,
			u.updated_at
		FROM support_conversations sc
		JOIN users u ON u.id = sc.user_id
	`
}

func baseSupportMessageQuery() string {
	return `
		SELECT
			id,
			conversation_id,
			sender_id,
			sender_role,
			body,
			created_at
		FROM support_messages
	`
}

func scanSupportConversations(rows *sql.Rows) ([]models.SupportConversation, error) {
	conversations := make([]models.SupportConversation, 0)

	for rows.Next() {
		var conversation models.SupportConversation
		var user models.User
		var lastMessageAt sql.NullTime

		if err := rows.Scan(
			&conversation.ID,
			&conversation.UserID,
			&conversation.Status,
			&lastMessageAt,
			&conversation.CreatedAt,
			&conversation.UpdatedAt,
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Name,
			&user.Email,
			&user.Rating,
			&user.RatingCount,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if lastMessageAt.Valid {
			conversation.LastMessageAt = &lastMessageAt.Time
		}
		conversation.User = &user
		conversation.Messages = []models.SupportMessage{}
		conversations = append(conversations, conversation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func scanSupportMessages(rows *sql.Rows) ([]models.SupportMessage, error) {
	messages := make([]models.SupportMessage, 0)

	for rows.Next() {
		var message models.SupportMessage
		if err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.SenderID,
			&message.SenderRole,
			&message.Body,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}

		message.Attachments = []models.SupportAttachment{}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func supportSenderRoleOrDefault(role string) string {
	if strings.EqualFold(strings.TrimSpace(role), models.SupportSenderAdmin) {
		return models.SupportSenderAdmin
	}

	return models.SupportSenderUser
}

func normalizeSupportConversationStatus(status string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case models.SupportConversationStatusOpen:
		return models.SupportConversationStatusOpen, true
	case models.SupportConversationStatusClosed:
		return models.SupportConversationStatusClosed, true
	default:
		return "", false
	}
}
