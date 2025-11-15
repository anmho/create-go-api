package posts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresPostTable is a repository for PostgreSQL operations on posts
type PostgresPostTable struct {
	db *pgxpool.Pool
}

// NewPostgresPostTable creates a new posts table repository and tests the connection
func NewPostgresPostTable(ctx context.Context, db *pgxpool.Pool) (*PostgresPostTable, error) {
	// Test connection
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return &PostgresPostTable{
		db: db,
	}, nil
}

func (t *PostgresPostTable) PutPost(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (id, user_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			updated_at = EXCLUDED.updated_at`

	_, err := t.db.Exec(ctx, query,
		post.ID, post.UserID, post.Title, post.Content, post.CreatedAt, post.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save post: %w", err)
	}
	return nil
}

// ListPostsByUserID returns all posts authored by the user with id userID
func (t *PostgresPostTable) ListPostsByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := t.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, nil
}

// GetPostByID retrieves a post by its ID
func (t *PostgresPostTable) GetPostByID(ctx context.Context, postID uuid.UUID) (*Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM posts
		WHERE id = $1`

	var post Post
	err := t.db.QueryRow(ctx, query, postID).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &post, nil
}

// DeletePost removes a post by its ID
func (t *PostgresPostTable) DeletePost(ctx context.Context, postID uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := t.db.Exec(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrPostNotFound
	}

	return nil
}

