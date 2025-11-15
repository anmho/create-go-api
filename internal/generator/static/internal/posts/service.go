package posts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Service defines the interface for post business logic
type Service interface {
	CreatePost(ctx context.Context, userID uuid.UUID, title, content string) (*Post, error)
	GetPost(ctx context.Context, postID uuid.UUID) (*Post, error)
	ListUserPosts(ctx context.Context, userID uuid.UUID) ([]Post, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, title, content string) (*Post, error)
	DeletePost(ctx context.Context, postID uuid.UUID) error
}

// service implements the Service interface
type service struct {
	postTable PostTable
}

// NewService creates a new posts service
func NewService(postTable PostTable) Service {
	return &service{postTable: postTable}
}

// CreatePost creates a new post
func (s *service) CreatePost(ctx context.Context, userID uuid.UUID, title, content string) (*Post, error) {
	post := NewPost(userID, title, content)
	if err := s.postTable.PutPost(ctx, post); err != nil {
		slog.ErrorContext(ctx, "Service: failed to create post", "error", err, "user_id", userID, "title", title)
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return post, nil
}

// GetPost retrieves a post by its ID
func (s *service) GetPost(ctx context.Context, postID uuid.UUID) (*Post, error) {
	post, err := s.postTable.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			slog.WarnContext(ctx, "Service: post not found", "post_id", postID)
		} else {
			slog.ErrorContext(ctx, "Service: failed to get post", "error", err, "post_id", postID)
		}
		return nil, fmt.Errorf("failed to get post by ID %v: %w", postID, err)
	}
	return post, nil
}

// ListUserPosts lists all posts for a given user
func (s *service) ListUserPosts(ctx context.Context, userID uuid.UUID) ([]Post, error) {
	posts, err := s.postTable.ListPostsByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Service: failed to list posts", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to list posts for user %s: %w", userID, err)
	}
	return posts, nil
}

// UpdatePost updates an existing post
func (s *service) UpdatePost(ctx context.Context, postID uuid.UUID, title, content string) (*Post, error) {
	existingPost, err := s.postTable.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			slog.WarnContext(ctx, "Service: post not found for update", "post_id", postID)
		} else {
			slog.ErrorContext(ctx, "Service: failed to find post to update", "error", err, "post_id", postID)
		}
		return nil, fmt.Errorf("failed to find post to update with ID %v: %w", postID, err)
	}

	// Update fields if provided
	if title != "" {
		existingPost.Title = title
	}
	if content != "" {
		existingPost.Content = content
	}
	existingPost.UpdatedAt = time.Now()

	if err := s.postTable.PutPost(ctx, existingPost); err != nil {
		slog.ErrorContext(ctx, "Service: failed to update post", "error", err, "post_id", postID)
		return nil, fmt.Errorf("failed to update post with ID %v: %w", postID, err)
	}
	return existingPost, nil
}

// DeletePost deletes a post by its ID
func (s *service) DeletePost(ctx context.Context, postID uuid.UUID) error {
	if err := s.postTable.DeletePost(ctx, postID); err != nil {
		if errors.Is(err, ErrPostNotFound) {
			slog.WarnContext(ctx, "Service: post not found for delete", "post_id", postID)
		} else {
			slog.ErrorContext(ctx, "Service: failed to delete post", "error", err, "post_id", postID)
		}
		return fmt.Errorf("failed to delete post with ID %v: %w", postID, err)
	}
	return nil
}

