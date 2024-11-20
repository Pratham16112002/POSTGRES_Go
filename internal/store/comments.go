package store

import (
	"context"
	"database/sql"
)

type CommentStore struct {
	db *sql.DB
}

type Comment struct {
	ID         int64  `json:"id"`
	PostID     int64  `json:"post_id"`
	UserID     int64  `json:"user_id"`
	Content    string `json:"content"`
	Created_At string `json:"created_at"`
	Updated_At string `json:"updated_at"`
	User       User   `json:"user"`
	Likes      int64  `json:"likes"`
}

func (c *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	query := `SELECT
	c.id , c.post_id, c.user_id , c.content,
	 c.created_at,c.updated_at , u.username , u.id 
FROM
	COMMENTS as  c
	JOIN USERS as u ON u.id = c.user_id
WHERE
	c.post_id = $1 
ORDER BY
	c.created_at DESC;`
	rows, err := c.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := []Comment{}
	for rows.Next() {
		var comment_row Comment
		comment_row.User = User{}
		err := rows.Scan(&comment_row.ID, &comment_row.PostID, &comment_row.UserID, &comment_row.Content,
			&comment_row.Created_At, &comment_row.Updated_At, &comment_row.User.Username, &comment_row.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment_row)
	}
	return comments, nil
}

func (c *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `INSERT INTO COMMENTS (post_id,user_id,content) VALUES( $1 , $2 , $3 ) RETURNING id`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := c.db.QueryRowContext(ctx, query, comment.PostID, comment.UserID, comment.Content).Scan(&comment.ID)
	if err != nil {
		return err
	}
	return nil
}
