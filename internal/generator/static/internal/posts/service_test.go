//go:build ignore

package posts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		postTable PostTable
	}{
		{
			name:      "creates service with table",
			postTable: NewMockPostTable(t),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.postTable)
			assert.NotNil(t, service)
		})
	}
}

func TestService_CreatePost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		userID       uuid.UUID
		title        string
		content      string
		setupMock    func(*MockPostTable)
		expectedErr  bool
		expectedPost *Post
	}{
		{
			name:    "successful creation",
			userID:  uuid.New(),
			title:   "Test Post",
			content: "Test Content",
			setupMock: func(m *MockPostTable) {
				m.On("PutPost", mock.Anything, mock.MatchedBy(func(post *Post) bool {
					return post.Title == "Test Post" && post.Content == "Test Content"
				})).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:    "table error",
			userID:  uuid.New(),
			title:   "Test Post",
			content: "Test Content",
			setupMock: func(m *MockPostTable) {
				m.On("PutPost", mock.Anything, mock.Anything).Return(errors.New("table error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTable := NewMockPostTable(t)
			tt.setupMock(mockTable)
			service := NewService(mockTable)

			post, err := service.CreatePost(context.Background(), tt.userID, tt.title, tt.content)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tt.userID, post.UserID)
				assert.Equal(t, tt.title, post.Title)
				assert.Equal(t, tt.content, post.Content)
				assert.NotEqual(t, uuid.Nil, post.ID)
			}
			mockTable.AssertExpectations(t)
		})
	}
}

func TestService_GetPost(t *testing.T) {
	t.Parallel()

	postID := uuid.New()
	userID := uuid.New()
	expectedPost := &Post{
		ID:        postID,
		UserID:    userID,
		Title:     "Test Post",
		Content:   "Test Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name         string
		postID       uuid.UUID
		setupMock    func(*MockPostTable)
		expectedErr  bool
		expectedPost *Post
	}{
		{
			name:   "successful retrieval",
			postID: postID,
			setupMock: func(m *MockPostTable) {
				m.On("GetPostByID", mock.Anything, postID).Return(expectedPost, nil)
			},
			expectedErr:  false,
			expectedPost: expectedPost,
		},
		{
			name:   "post not found",
			postID: postID,
			setupMock: func(m *MockPostTable) {
				m.On("GetPostByID", mock.Anything, postID).Return(nil, ErrPostNotFound)
			},
			expectedErr:  true,
			expectedPost: nil,
		},
		{
			name:   "table error",
			postID: postID,
			setupMock: func(m *MockPostTable) {
				m.On("GetPostByID", mock.Anything, postID).Return(nil, errors.New("table error"))
			},
			expectedErr:  true,
			expectedPost: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTable := NewMockPostTable(t)
			tt.setupMock(mockTable)
			service := NewService(mockTable)

			post, err := service.GetPost(context.Background(), tt.postID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPost, post)
			}
			mockTable.AssertExpectations(t)
		})
	}
}

func TestService_ListUserPosts(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expectedPosts := []Post{
		{ID: uuid.New(), UserID: userID, Title: "Post 1", Content: "Content 1"},
		{ID: uuid.New(), UserID: userID, Title: "Post 2", Content: "Content 2"},
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMock     func(*MockPostTable)
		expectedErr   bool
		expectedPosts []Post
	}{
		{
			name:   "successful list",
			userID: userID,
			setupMock: func(m *MockPostTable) {
				m.On("ListPostsByUserID", mock.Anything, userID).Return(expectedPosts, nil)
			},
			expectedErr:   false,
			expectedPosts: expectedPosts,
		},
		{
			name:   "empty list",
			userID: userID,
			setupMock: func(m *MockPostTable) {
				m.On("ListPostsByUserID", mock.Anything, userID).Return([]Post{}, nil)
			},
			expectedErr:   false,
			expectedPosts: []Post{},
		},
		{
			name:   "table error",
			userID: userID,
			setupMock: func(m *MockPostTable) {
				m.On("ListPostsByUserID", mock.Anything, userID).Return(nil, errors.New("table error"))
			},
			expectedErr:   true,
			expectedPosts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTable := NewMockPostTable(t)
			tt.setupMock(mockTable)
			service := NewService(mockTable)

			posts, err := service.ListUserPosts(context.Background(), tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPosts, posts)
			}
			mockTable.AssertExpectations(t)
		})
	}
}

func TestService_UpdatePost(t *testing.T) {
	t.Parallel()

	postID := uuid.New()
	userID := uuid.New()
	existingPost := &Post{
		ID:        postID,
		UserID:    userID,
		Title:     "Old Title",
		Content:   "Old Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name        string
		postID      uuid.UUID
		title       string
		content     string
		setupMock   func(*MockPostTable)
		expectedErr bool
	}{
		{
			name:    "successful update",
			postID:  postID,
			title:   "New Title",
			content: "New Content",
			setupMock: func(m *MockPostTable) {
				m.On("GetPostByID", mock.Anything, postID).Return(existingPost, nil)
				m.On("PutPost", mock.Anything, mock.MatchedBy(func(post *Post) bool {
					return post.ID == postID && post.Title == "New Title" && post.Content == "New Content"
				})).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:    "post not found",
			postID:  postID,
			title:   "New Title",
			content: "New Content",
			setupMock: func(m *MockPostTable) {
				m.On("GetPostByID", mock.Anything, postID).Return(nil, ErrPostNotFound)
			},
			expectedErr: true,
		},
		{
			name:    "table error on get",
			postID:  postID,
			title:   "New Title",
			content: "New Content",
			setupMock: func(m *MockPostTable) {
				m.On("GetPostByID", mock.Anything, postID).Return(nil, errors.New("table error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTable := NewMockPostTable(t)
			tt.setupMock(mockTable)
			service := NewService(mockTable)

			post, err := service.UpdatePost(context.Background(), tt.postID, tt.title, tt.content)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tt.title, post.Title)
				assert.Equal(t, tt.content, post.Content)
			}
			mockTable.AssertExpectations(t)
		})
	}
}

func TestService_DeletePost(t *testing.T) {
	t.Parallel()

	postID := uuid.New()

	tests := []struct {
		name        string
		postID      uuid.UUID
		setupMock   func(*MockPostTable)
		expectedErr bool
	}{
		{
			name:   "successful deletion",
			postID: postID,
			setupMock: func(m *MockPostTable) {
				m.On("DeletePost", mock.Anything, postID).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:   "post not found",
			postID: postID,
			setupMock: func(m *MockPostTable) {
				m.On("DeletePost", mock.Anything, postID).Return(ErrPostNotFound)
			},
			expectedErr: true,
		},
		{
			name:   "table error",
			postID: postID,
			setupMock: func(m *MockPostTable) {
				m.On("DeletePost", mock.Anything, postID).Return(errors.New("table error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTable := NewMockPostTable(t)
			tt.setupMock(mockTable)
			service := NewService(mockTable)

			err := service.DeletePost(context.Background(), tt.postID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockTable.AssertExpectations(t)
		})
	}
}

