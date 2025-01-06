package store

import (
	"Blog/internal/store/paginate"
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
	RoleID    int64        `json:"role_id"`
	Role      Role         `json:"role"`
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

func (p *PasswordType) Compare(pass string) error {
	if err := bcrypt.CompareHashAndPassword(p.hash, []byte(pass)); err != nil {
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
	query := `INSERT INTO users (Username,Email,Password,role_id) 
			VALUES ($1 , $2 , $3 , (SELECT id FROM roles WHERE name = $4) ) RETURNING id,created_at`

	role := user.Role.Name
	if role == "" {
		role = "user"
	}
	err := tx.QueryRowContext(ctx, query, user.Username, user.Email,
		user.Password.hash, role).Scan(&user.ID,
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

func (s *UserStore) deleteUser(ctx context.Context, tx *sql.Tx, userId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `DELETE FROM users where id = $1`
	sql_res, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}
	rows_affected, err := sql_res.RowsAffected()
	if err != nil {
		return err
	}
	if rows_affected == 0 {
		return ErrNotFound
	}
	return nil

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

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT id,username , email, password , created_at, is_active FROM users WHERE email = $1 AND is_active = true`
	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password.hash, &user.CreatedAt, &user.IsActive)
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

func (s *UserStore) GetUserById(ctx context.Context, UserId int64) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT users.id,username , email, created_at , is_active, role_id , roles.* FROM users JOIN roles ON (users.role_id = roles.id) WHERE users.id = $1`
	user := &User{}
	err := s.db.QueryRowContext(ctx, query, UserId).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.IsActive, &user.RoleID, &user.Role.ID, &user.Role.Name, &user.Role.Level, &user.Role.Description)
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

func (s *UserStore) Delete(ctx context.Context, userId int64) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// deleting the user invitation
		err := s.deleteUserInvitations(ctx, tx, userId)
		if err != nil {
			return err
		}

		err = s.deleteUser(ctx, tx, userId)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) SearchFriends(ctx context.Context, userId int64, friendQuery *paginate.FriendPaginateQuery) ([]User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT 
    u.id, 
    u.username,
    u.email,
    u.created_at
FROM 
    users AS u 
JOIN 
    roles AS r 
ON 
    u.role_id = r.id 
WHERE 
    r.name = $1 
AND 
    u.id NOT IN (
        SELECT 
            follower_id 
        FROM 
            followers 
        WHERE 
            user_id = $2 
    )
AND
	u.id != $2
LIMIT $3 OFFSET $4;
`
	rows, err := s.db.QueryContext(ctx, query, friendQuery.Role, userId, friendQuery.Limit, friendQuery.Offset)
	if err != nil {
		return nil, err
	}
	list := []User{}
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}
