package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserId    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	Version   int       `json:"version"`
	User      User      `json:"user"`
}

type PostWithMetaData struct {
	Post
	CommentCount int `json:"comment_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) GetById(ctx context.Context, id int64) (*Post, error) {
	query := `SELECT id,title,content,created_at,updated_at,tags,version from
		 posts where id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt, pq.Array(&post.Tags), &post.Version)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId int64) ([]PostWithMetaData, error) {
	query
}

func (s *PostStore) Delete(ctx context.Context, postID int64) error {
	query := `DELETE FROM posts where id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query_result, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	rows_affected, err := query_result.RowsAffected()
	if err != nil {
		return err
	}
	if rows_affected == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (content,title,user_id,tags) 
		values ($1 , $2 , $3 , $4)
		 RETURNING id,created_at , updated_at`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(&post.Tags)).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `UPDATE 
	posts SET title = $1 , content = $2 , version = version + 1 WHERE
	id = $3 AND version = $4 RETURNING version`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}
