package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64        `json:"id"`
	Username  string       `json:"username,omitempty"`
	Email     string       `json:"email,omitempty"`
	Password  PasswordType `json:"-"`
	CreatedAt string       `json:"created_at,omitempty"`
	IsActive  bool         `json:"is_active,omitempty"`
}

type PasswordType struct {
	text *string
	hash []byte
}

func (p *PasswordType) Set(password_txt string) error {
	var err error
	p.hash, err = bcrypt.GenerateFromPassword([]byte(password_txt), bcrypt.DefaultCost)
	p.text = &password_txt
	if err != nil {
		return err
	}
	return nil

}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `INSERT INTO users (Username,Email,Password) 
			VALUES ($1 , $2 , $3 ) RETURNING id,created_at`
	err := tx.QueryRowContext(ctx, query, user.Username, user.Email,
		user.Password.hash).Scan(&user.ID,
		&user.CreatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	// Transaction wrapper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}
		// create user invitation
		if err := s.createUserInvitation(ctx, tx, token, exp, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, UserId int64) error {
	query := `INSERT INTO user_invitations 
	(token,user_id,expiry) VALUES
		($1, $2, $3);`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, token, UserId, time.Now().Add(exp))
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) GetUserById(ctx context.Context, UserId int64) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT id,username , email, created_at FROM users WHERE ID = $1`
	user := &User{}
	err := s.db.QueryRowContext(ctx, query, UserId).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
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

func (s *UserStore) Activate(ctx context.Context, token string) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}
	return nil

}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `UPDATE users SET username = $1 , email = $2 , is_active = $3 WHERE id = $4`

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT u.id , u.username , u.email , u.created_at , u.is_active FROM users u JOIN user_invitations
	ui ON u.id = ui.user_id WHERE ui.token = $1 AND ui.expiry > $2`
	user := &User{}
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])
	err := tx.QueryRowContext(ctx, query, hashedToken, time.Now()).Scan(&user.ID,
		&user.Username, &user.Email,
		&user.CreatedAt, &user.IsActive)
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

func (s *UserStore) Delete(ctx context.Context)
