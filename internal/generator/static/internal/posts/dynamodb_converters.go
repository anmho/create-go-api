package posts

import (
	"time"

	"github.com/google/uuid"
)

// DynamoDBPostStorageModel represents the DynamoDB storage format for a Post
type DynamoDBPostStorageModel struct {
	UserID    string `dynamodbav:"UserID"`
	CreatedAt int64  `dynamodbav:"CreatedAt"`
	PostID    string `dynamodbav:"PostID"`
	Title     string `dynamodbav:"Title"`
	Content   string `dynamodbav:"Content"`
	UpdatedAt int64  `dynamodbav:"UpdatedAt"`
}

// DynamoDBPostToStorage converts a Post model to a DynamoDBPostStorageModel
func DynamoDBPostToStorage(post *Post) *DynamoDBPostStorageModel {
	return &DynamoDBPostStorageModel{
		UserID:    post.UserID.String(),
		CreatedAt: post.CreatedAt.UnixMilli(),
		PostID:    post.ID.String(),
		Title:     post.Title,
		Content:   post.Content,
		UpdatedAt: post.UpdatedAt.UnixMilli(),
	}
}

// DynamoDBStorageToPost converts a DynamoDBPostStorageModel to a Post model
func DynamoDBStorageToPost(storage *DynamoDBPostStorageModel) (*Post, error) {
	userID, err := uuid.Parse(storage.UserID)
	if err != nil {
		return nil, err
	}

	postID, err := uuid.Parse(storage.PostID)
	if err != nil {
		return nil, err
	}

	return &Post{
		ID:        postID,
		UserID:    userID,
		Title:     storage.Title,
		Content:   storage.Content,
		CreatedAt: time.UnixMilli(storage.CreatedAt),
		UpdatedAt: time.UnixMilli(storage.UpdatedAt),
	}, nil
}

