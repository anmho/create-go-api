//go:build ignore

package api

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/acme/postservice/internal/posts"
	postsv1 "github.com/acme/postservice/internal/protos/gen/posts/v1"
	postsv1connect "github.com/acme/postservice/internal/protos/gen/posts/v1/postsv1connect"
)

// PostServiceHandler implements the gRPC PostService
type PostServiceHandler struct {
	postsv1connect.UnimplementedPostServiceHandler
	service posts.Service
}

// NewPostServiceHandler creates a new gRPC handler for posts
func NewPostServiceHandler(service posts.Service) *PostServiceHandler {
	return &PostServiceHandler{
		service: service,
	}
}

// CreatePost handles post creation requests
func (h *PostServiceHandler) CreatePost(
	ctx context.Context,
	req *postsv1.CreatePostRequest,
) (*postsv1.CreatePostResponse, error) {
	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid user_id", "error", err, "user_id", req.UserId)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid user_id"))
	}

	// Create post
	post, err := h.service.CreatePost(ctx, userID, req.Title, req.Content)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create post", "error", err, "user_id", userID)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create post"))
	}

	// Convert to proto
	protoPost := posts.PostToProto(post)

	return &postsv1.CreatePostResponse{
		Post: protoPost,
	}, nil
}

// GetPost retrieves a post by ID
func (h *PostServiceHandler) GetPost(
	ctx context.Context,
	req *postsv1.GetPostRequest,
) (*postsv1.GetPostResponse, error) {
	// Parse post ID
	postID, err := uuid.Parse(req.PostId)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid post_id", "error", err, "post_id", req.PostId)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid post_id"))
	}

	// Get post
	post, err := h.service.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, posts.ErrPostNotFound) {
			slog.WarnContext(ctx, "Post not found", "post_id", postID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
		}
		slog.ErrorContext(ctx, "Failed to get post", "error", err, "post_id", postID)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get post"))
	}

	// Convert to proto
	protoPost := posts.PostToProto(post)

	return &postsv1.GetPostResponse{
		Post: protoPost,
	}, nil
}

// ListPosts retrieves all posts for a user
func (h *PostServiceHandler) ListPosts(
	ctx context.Context,
	req *postsv1.ListPostsRequest,
) (*postsv1.ListPostsResponse, error) {
	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid user_id", "error", err, "user_id", req.UserId)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid user_id"))
	}

	// List posts
	postsList, err := h.service.ListUserPosts(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list posts", "error", err, "user_id", userID)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to list posts"))
	}

	// Convert to proto
	protoPosts := make([]*postsv1.Post, 0, len(postsList))
	for i := range postsList {
		protoPosts = append(protoPosts, posts.PostToProto(&postsList[i]))
	}

	return &postsv1.ListPostsResponse{
		Posts: protoPosts,
	}, nil
}

// UpdatePost updates an existing post
func (h *PostServiceHandler) UpdatePost(
	ctx context.Context,
	req *postsv1.UpdatePostRequest,
) (*postsv1.UpdatePostResponse, error) {
	// Parse post ID
	postID, err := uuid.Parse(req.PostId)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid post_id", "error", err, "post_id", req.PostId)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid post_id"))
	}

	// Prepare update values
	title := ""
	if req.Title != nil {
		title = *req.Title
	}
	content := ""
	if req.Content != nil {
		content = *req.Content
	}

	// Update post
	post, err := h.service.UpdatePost(ctx, postID, title, content)
	if err != nil {
		if errors.Is(err, posts.ErrPostNotFound) {
			slog.WarnContext(ctx, "Post not found for update", "post_id", postID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
		}
		slog.ErrorContext(ctx, "Failed to update post", "error", err, "post_id", postID)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to update post"))
	}

	// Convert to proto
	protoPost := posts.PostToProto(post)

	return &postsv1.UpdatePostResponse{
		Post: protoPost,
	}, nil
}

// DeletePost deletes a post by ID
func (h *PostServiceHandler) DeletePost(
	ctx context.Context,
	req *postsv1.DeletePostRequest,
) (*postsv1.DeletePostResponse, error) {
	// Parse post ID
	postID, err := uuid.Parse(req.PostId)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid post_id", "error", err, "post_id", req.PostId)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid post_id"))
	}

	// Delete post
	err = h.service.DeletePost(ctx, postID)
	if err != nil {
		if errors.Is(err, posts.ErrPostNotFound) {
			slog.WarnContext(ctx, "Post not found for delete", "post_id", postID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
		}
		slog.ErrorContext(ctx, "Failed to delete post", "error", err, "post_id", postID)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to delete post"))
	}

	return &postsv1.DeletePostResponse{
		Message: "Post deleted successfully",
	}, nil
}
