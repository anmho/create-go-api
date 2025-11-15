package posts

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockery

// PostTable defines the interface for post data operations
// This interface is implemented by both Postgres and DynamoDB table implementations
type PostTable interface {
	PutPost(ctx context.Context, post *Post) error
	GetPostByID(ctx context.Context, postID uuid.UUID) (*Post, error)
	ListPostsByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error)
	DeletePost(ctx context.Context, postID uuid.UUID) error
}

