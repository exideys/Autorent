package models

import "time"

const (
	SupportSenderUser  = "user"
	SupportSenderAdmin = "admin"

	SupportConversationStatusOpen   = "open"
	SupportConversationStatusClosed = "closed"
)

type SupportConversation struct {
	ID            int64            `json:"id"`
	UserID        int64            `json:"user_id"`
	User          *User            `json:"user,omitempty"`
	Status        string           `json:"status"`
	LastMessageAt *time.Time       `json:"last_message_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Messages      []SupportMessage `json:"messages,omitempty"`
}

type SupportMessage struct {
	ID             int64               `json:"id"`
	ConversationID int64               `json:"conversation_id"`
	SenderID       int64               `json:"sender_id"`
	SenderRole     string              `json:"sender_role"`
	Body           string              `json:"body"`
	CreatedAt      time.Time           `json:"created_at"`
	Attachments    []SupportAttachment `json:"attachments,omitempty"`
}

type SupportAttachment struct {
	ID          int64     `json:"id"`
	MessageID   int64     `json:"message_id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	DriveFileID string    `json:"drive_file_id"`
	FileURL     string    `json:"file_url"`
	DriveURL    *string   `json:"drive_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type SupportAttachmentInput struct {
	FileName    string
	ContentType string
	FileSize    int64
	DriveFileID string
	FileURL     string
	DriveURL    *string
}
