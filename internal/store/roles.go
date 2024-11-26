package store

import (
	"context"
	"database/sql"
	"errors"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int64  `json:"level"`
	Description string `json:"description"`
}

type RoleStore struct {
	db *sql.DB
}

func (r *RoleStore) GetRoleByName(ctx context.Context, role_name string) (*Role, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	query := `SELECT id, name , level , description FROM roles WHERE name = $1`
	role := &Role{}
	err := r.db.QueryRowContext(ctx, query, role_name).Scan(&role.ID, &role.Name, &role.Level, &role.Description)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return role, err
}
