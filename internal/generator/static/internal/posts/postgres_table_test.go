package posts

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresPostTable_Serialization(t *testing.T) {
	ctx := context.Background()

	// Start Postgres container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(ctx))
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	// Create table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS posts (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create table instance
	table, err := NewPostgresPostTable(ctx, pool)
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
				// Use UTC time to avoid timezone conversion issues with PostgreSQL
				utcNow := now.UTC()
				post := &Post{
					ID:        postID,
					UserID:    userID,
					Title:     "Test Post",
					Content:   "Test Content",
					CreatedAt: utcNow,
					UpdatedAt: utcNow,
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
				// Compare times - both should be in UTC
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

