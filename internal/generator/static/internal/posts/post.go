package posts

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a blog post or similar content
type Post struct {
	ID        uuid.UUID `json:"id" dynamodbav:"PostID" db:"id"`
	UserID    uuid.UUID `json:"user_id" dynamodbav:"UserID" db:"user_id"`
	Title     string    `json:"title" dynamodbav:"Title" db:"title"`
	Content   string    `json:"content" dynamodbav:"Content" db:"content"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"CreatedAt" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"UpdatedAt" db:"updated_at"`
}

// NewPost creates a new Post instance
func NewPost(userID uuid.UUID, title, content string) *Post {
	now := time.Now()
	return &Post{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

