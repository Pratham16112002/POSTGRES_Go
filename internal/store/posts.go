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
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Comments  []Comment `json:"comments,omitempty"`
	Version   int       `json:"version"`
	User      User      `json:"user"`
}

type PostWithMetaData struct {
	Post
	CommentCount int64 `json:"comment_count"`
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId int64, pageQuery PaginatedFeedQuery) ([]PostWithMetaData, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT
	p.id,
	p.user_id,
	p.title,
	p.content,
	p.created_at,
	p.updated_at,
	p.version,
	p.tags,
	u.username,
	COUNT(c.id) AS comment_count
FROM
	posts p
	LEFT JOIN comments c ON p.id = c.post_id
	LEFT JOIN users u ON p.user_id = u.id
	LEFT JOIN followers f ON f.follower_id = p.user_id
WHERE
	(f.user_id = $1 
	OR p.user_id = $1
	)
	AND
	(
	p.title ILIKE  '%' || $2 || '%'
	OR
	p.content ILIKE  '%' || $2 || '%'
	)
	AND
	( p.tags @> $3 OR $3 = '{}' )
GROUP BY
	p.id, u.username
ORDER BY
	p.created_at ` + pageQuery.Sort + ` LIMIT $4 OFFSET $5;`

	rows, err := s.db.QueryContext(ctx, query, userId,
		pageQuery.Search, pq.Array(pageQuery.Tags),
		pageQuery.Limit,
		pageQuery.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	feed := []PostWithMetaData{}
	for rows.Next() {
		var p PostWithMetaData
		err := rows.Scan(&p.ID, &p.UserId, &p.Title,
			&p.Content, &p.CreatedAt, &p.UpdatedAt, &p.Version,
			pq.Array(&p.Tags),
			&p.User.Username, &p.CommentCount)
		if err != nil {
			return nil, err
		}
		feed = append(feed, p)
	}
	return feed, nil
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
