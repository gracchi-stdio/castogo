package sqlite

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/sqlcsqlite"
)

type UserRepo struct {
	q *sqlcsqlite.Queries
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{q: sqlcsqlite.New(db)}
}

func (ur UserRepo) GetByID(ctx context.Context, id [16]byte) (*domain.User, error) {
	row, err := ur.q.GetUserByID(ctx, uuidToHex(id))
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := ur.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := ur.q.CreateUser(ctx, sqlcsqlite.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := ur.q.UpdateUser(ctx, sqlcsqlite.UpdateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		ID:           uuidToHex(user.ID),
	})
	if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func (ur UserRepo) Delete(ctx context.Context, id [16]byte) error {
	return ur.q.DeleteUser(ctx, uuidToHex(id))
}

// uuidToHex returns the lowercase hexadecimal encoding of id as a 32-character string.
func uuidToHex(id [16]byte) string {
	return hex.EncodeToString(id[:])
}

// hexToUUID decodes a hexadecimal string into a 16-byte UUID.
// It returns an error if s is not valid hex or does not decode to exactly 16 bytes.
func hexToUUID(s string) ([16]byte, error) {
	var id [16]byte
	if _, err := hex.Decode(id[:], []byte(s)); err != nil {
		return [16]byte{}, fmt.Errorf("parse user id %q: %w", s, err)
	}
	return id, nil
}

const sqliteTimeLayout = "2006-01-02 15:04:05"

func sqliteTime(s *string) time.Time {
	if s == nil {
		return time.Time{}
	}
	t, _ := time.Parse(sqliteTimeLayout, *s)

	return t
}

func userFromRow(u sqlcsqlite.User) (*domain.User, error) {
	id, err := hexToUUID(u.ID)
	if err != nil {
		return nil, err
	}
	created, err := time.Parse(sqliteTimeLayout, u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse user created_at %q: %w", u.CreatedAt, err)
	}

	return &domain.User{
		ID:           id,
		PasswordHash: u.PasswordHash,
		CreatedAt:    created,
		UpdatedAt:    sqliteTime(u.UpdatedAt),
	}, nil
}
