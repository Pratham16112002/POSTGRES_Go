package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}
type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `INSERT INTO users (Username,Email,Password) 
			VALUES ($1 , $2 , $3 ) RETURNING id,created_at`
	err := s.db.QueryRowContext(ctx, query, user.Username, user.Email,
		user.CreatedAt).Scan(&user.ID,
		&user.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) GetUserById(ctx context.Context, userId int64) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT id,username , email, created_at FROM 	users WHERE ID = $1`
	user := &User{}
	err := s.db.QueryRowContext(ctx, query, userId).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}

	}
	return user, nil
}
