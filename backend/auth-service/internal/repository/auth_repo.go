package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(ctx context.Context, email, passwordHash, name string) (uuid.UUID, error) {
	var id uuid.UUID
	query := `INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, email, passwordHash, name).Scan(&id)
	return id, err
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, name FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	return &user, err
}
