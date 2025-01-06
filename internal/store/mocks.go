package store

import (
	"Blog/internal/store/paginate"
	"context"
	"database/sql"
	"time"
)

type MockUserStore struct{}

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, u *User) error {
	return nil
}

func (m *MockUserStore) GetUserById(ctx context.Context, userID int64) (*User, error) {
	return &User{ID: userID}, nil
}

func (m *MockUserStore) GetUserByEmail(context.Context, string) (*User, error) {
	return &User{}, nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, t string) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockUserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	return nil
}

func (m *MockUserStore) SearchFriends(ctx context.Context, userId int64, friendQuery *paginate.FriendPaginateQuery) ([]User, error) {
	return nil, nil
}
