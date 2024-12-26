package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Follower struct {
	FollweeID int64  `json:"user_id"`
	FollwerID int64  `json:"follower_id"`
	CreatedAt string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followeeId int64, followerId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	fmt.Println(followeeId, followerId)
	query := `INSERT INTO followers(user_id,follower_id) VALUES( $1 , $2 )`
	_, err := s.db.ExecContext(ctx, query, followeeId, followerId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return err
}

func (s *FollowerStore) Unfollow(ctx context.Context, followeeId int64, followerId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `DELETE FROM followers WHERE user_id = $1 AND follower_id = $2`
	_, err := s.db.ExecContext(ctx, query, followeeId, followerId)
	return err
}
