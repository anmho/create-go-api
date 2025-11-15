package posts

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDynamoDBPostTable_Serialization(t *testing.T) {
	ctx := context.Background()

	// Start DynamoDB Local container
	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local:latest",
		ExposedPorts: []string{"8000/tcp"},
		WaitingFor:   wait.ForListeningPort("8000/tcp").WithStartupTimeout(30 * time.Second),
	}

	dynamoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, dynamoContainer.Terminate(ctx))
	}()

	// Get endpoint
	endpoint, err := dynamoContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	// Create DynamoDB client with dummy credentials for local DynamoDB
	cfg := aws.Config{
		Region:       "us-east-1",
		BaseEndpoint: aws.String("http://" + endpoint),
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("local", "local", "")),
	}
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Create table instance (table will be created automatically if it doesn't exist)
	table, err := NewDynamoDBPostTable(ctx, dynamoClient)
	require.NoError(t, err)

	// Wait for table to be active (in case it was just created)
	waiter := dynamodb.NewTableExistsWaiter(dynamoClient)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(PostTableName),
	}, 30*time.Second)
	require.NoError(t, err)

	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name string
		fn   func(t *testing.T, table PostTable, userID uuid.UUID, now time.Time)
	}{
		{
			name: "PutPost and GetPostByID - serialization roundtrip",
			fn: func(t *testing.T, table PostTable, userID uuid.UUID, now time.Time) {
				postID := uuid.New()
				post := &Post{
					ID:        postID,
					UserID:    userID,
					Title:     "Test Post",
					Content:   "Test Content",
					CreatedAt: now,
					UpdatedAt: now,
				}

				// Put post
				err := table.PutPost(ctx, post)
				require.NoError(t, err)

				// Get post back
				retrieved, err := table.GetPostByID(ctx, postID)
				require.NoError(t, err)
				require.NotNil(t, retrieved)

				// Verify serialization - all fields should match
				assert.Equal(t, post.ID, retrieved.ID)
				assert.Equal(t, post.UserID, retrieved.UserID)
				assert.Equal(t, post.Title, retrieved.Title)
				assert.Equal(t, post.Content, retrieved.Content)
				assert.WithinDuration(t, post.CreatedAt, retrieved.CreatedAt, time.Second)
				assert.WithinDuration(t, post.UpdatedAt, retrieved.UpdatedAt, time.Second)
			},
		},
		{
			name: "ListPostsByUserID - serialization",
			fn: func(t *testing.T, table PostTable, userID uuid.UUID, now time.Time) {
				post1 := &Post{
					ID:        uuid.New(),
					UserID:    userID,
					Title:     "Post 1",
					Content:   "Content 1",
					CreatedAt: now.Add(-2 * time.Hour),
					UpdatedAt: now.Add(-2 * time.Hour),
				}
				post2 := &Post{
					ID:        uuid.New(),
					UserID:    userID,
					Title:     "Post 2",
					Content:   "Content 2",
					CreatedAt: now.Add(-1 * time.Hour),
					UpdatedAt: now.Add(-1 * time.Hour),
				}

				err := table.PutPost(ctx, post1)
				require.NoError(t, err)
				err = table.PutPost(ctx, post2)
				require.NoError(t, err)

				// List posts
				posts, err := table.ListPostsByUserID(ctx, userID)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(posts), 2)

				// Verify serialization for at least one post
				found := false
				for _, p := range posts {
					if p.ID == post1.ID {
						assert.Equal(t, post1.Title, p.Title)
						assert.Equal(t, post1.Content, p.Content)
						found = true
						break
					}
				}
				assert.True(t, found, "post1 should be in the list")
			},
		},
		{
			name: "DeletePost",
			fn: func(t *testing.T, table PostTable, userID uuid.UUID, now time.Time) {
				deletePostID := uuid.New()
				post := &Post{
					ID:        deletePostID,
					UserID:    userID,
					Title:     "To Delete",
					Content:   "Will be deleted",
					CreatedAt: now,
					UpdatedAt: now,
				}

				err := table.PutPost(ctx, post)
				require.NoError(t, err)

				// Delete post
				err = table.DeletePost(ctx, deletePostID)
				require.NoError(t, err)

				// Verify it's gone
				_, err = table.GetPostByID(ctx, deletePostID)
				assert.Error(t, err)
				assert.Equal(t, ErrPostNotFound, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, table, userID, now)
		})
	}
}

