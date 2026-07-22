package user

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		SELECT id, display_name, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var foundUser User
	err := repository.pool.QueryRow(ctx, query, id).Scan(
		&foundUser.ID,
		&foundUser.DisplayName,
		&foundUser.Email,
		&foundUser.Password,
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

func (repository *PostgresRepository) UpdatePassword(
	ctx context.Context,
	id uuid.UUID,
	passwordHash string,
) error {
	query := `
		UPDATE users
		SET password_hash = $2,
			updated_at = NOW()
		WHERE id = $1
	`

	commandTag, err := repository.pool.Exec(ctx, query, id, passwordHash)
	if err != nil {
		return fmt.Errorf("cannot update user password: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (repository *PostgresRepository) CreatePasswordResetToken(
	ctx context.Context,
	resetToken PasswordResetToken,
) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password reset token transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(
		ctx,
		`
			UPDATE password_reset_tokens
			SET used_at = $2
			WHERE user_id = $1
				AND used_at IS NULL
		`,
		resetToken.UserID,
		resetToken.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("cannot invalidate previous password reset tokens: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
			INSERT INTO password_reset_tokens (
				id,
				user_id,
				token_hash,
				expires_at,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5)
		`,
		resetToken.ID,
		resetToken.UserID,
		resetToken.TokenHash,
		resetToken.ExpiresAt,
		resetToken.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("cannot create password reset token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password reset token transaction: %w", err)
	}

	return nil
}

func (repository *PostgresRepository) ResetPasswordWithToken(
	ctx context.Context,
	tokenHash string,
	passwordHash string,
	now time.Time,
) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password reset transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	err = tx.QueryRow(
		ctx,
		`
			SELECT password_reset_tokens.user_id
			FROM password_reset_tokens
			JOIN users ON users.id = password_reset_tokens.user_id
			WHERE password_reset_tokens.token_hash = $1
				AND password_reset_tokens.used_at IS NULL
				AND password_reset_tokens.expires_at > $2
				AND users.status = $3
			FOR UPDATE OF password_reset_tokens
		`,
		tokenHash,
		now,
		StatusActive,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidPasswordResetToken
		}

		return fmt.Errorf("cannot find password reset token: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
			UPDATE users
			SET password_hash = $2,
				updated_at = $3
			WHERE id = $1
		`,
		userID,
		passwordHash,
		now,
	)
	if err != nil {
		return fmt.Errorf("cannot reset user password: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
			UPDATE password_reset_tokens
			SET used_at = $2
			WHERE user_id = $1
				AND used_at IS NULL
		`,
		userID,
		now,
	)
	if err != nil {
		return fmt.Errorf("cannot consume password reset tokens: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password reset transaction: %w", err)
	}

	return nil
}
