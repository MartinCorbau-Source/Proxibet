package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolationCode = "23505"

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (repository *PostgresRepository) Create(
	ctx context.Context,
	newUser User,
) (User, error) {
	query := `
		INSERT INTO users (
		id,
		display_name,
		email,
		password_hash,
		status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, display_name, email, password_hash, status, created_at, updated_at
	`
	var createdUser User
	err := repository.pool.QueryRow(
		ctx,
		query,
		newUser.ID,
		newUser.DisplayName,
		newUser.Email,
		newUser.Password,
		newUser.Status,
	).Scan(
		&createdUser.ID,
		&createdUser.DisplayName,
		&createdUser.Email,
		&createdUser.Password,
		&createdUser.Status,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return User{}, ErrEmailAlreadyInUse
		}
		return User{}, fmt.Errorf("cannot create user: %w", err)
	}
	return createdUser, nil
}

func (repository *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (User, error) {
	query := `
		SELECT id, display_name, email, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var foundUser User
	err := repository.pool.QueryRow(ctx, query, id).Scan(
		&foundUser.ID,
		&foundUser.DisplayName,
		&foundUser.Email,
		&foundUser.Status,
		&foundUser.CreatedAt,
		&foundUser.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("cannot find user: %w", err)
	}

	return foundUser, nil
}

func (repository *PostgresRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	query := `
		SELECT id, display_name, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
	var foundUser User
	err := repository.pool.QueryRow(ctx, query, email).Scan(
		&foundUser.ID,
		&foundUser.DisplayName,
		&foundUser.Email,
		&foundUser.Password,
		&foundUser.Status,
		&foundUser.CreatedAt,
		&foundUser.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("cannot find user: %w", err)
	}

	return foundUser, nil
}
