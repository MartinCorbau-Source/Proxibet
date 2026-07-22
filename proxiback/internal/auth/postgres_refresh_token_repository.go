package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(pool *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{
		pool: pool,
	}
}

func (repository *PostgresRefreshTokenRepository) Create(
	ctx context.Context,
	refreshToken RefreshToken,
) (RefreshToken, error) {
	query := `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, replaced_by_token_id, created_at
	`

	var createdRefreshToken RefreshToken
	err := repository.pool.QueryRow(
		ctx,
		query,
		refreshToken.ID,
		refreshToken.UserID,
		refreshToken.TokenHash,
		refreshToken.ExpiresAt,
	).Scan(
		&createdRefreshToken.ID,
		&createdRefreshToken.UserID,
		&createdRefreshToken.TokenHash,
		&createdRefreshToken.ExpiresAt,
		&createdRefreshToken.RevokedAt,
		&createdRefreshToken.ReplacedByTokenID,
		&createdRefreshToken.CreatedAt,
	)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("cannot create refresh token: %w", err)
	}

	return createdRefreshToken, nil
}

func (repository *PostgresRefreshTokenRepository) Rotate(
	ctx context.Context,
	tokenHash string,
	nextRefreshToken RefreshToken,
	revokedAt time.Time,
) (RefreshToken, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return RefreshToken{}, fmt.Errorf("begin refresh token rotation: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	currentRefreshToken, err := findRefreshTokenForUpdate(ctx, tx, tokenHash)
	if err != nil {
		return RefreshToken{}, err
	}

	if currentRefreshToken.RevokedAt != nil {
		return RefreshToken{}, ErrRefreshTokenRevoked
	}

	if !currentRefreshToken.ExpiresAt.After(revokedAt) {
		if err := revokeRefreshToken(ctx, tx, currentRefreshToken, revokedAt, nil); err != nil {
			return RefreshToken{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return RefreshToken{}, fmt.Errorf("commit expired refresh token revocation: %w", err)
		}

		return RefreshToken{}, ErrRefreshTokenExpired
	}

	nextRefreshToken.UserID = currentRefreshToken.UserID
	createdRefreshToken, err := createRefreshToken(ctx, tx, nextRefreshToken)
	if err != nil {
		return RefreshToken{}, err
	}

	if err := revokeRefreshToken(ctx, tx, currentRefreshToken, revokedAt, &createdRefreshToken.ID); err != nil {
		return RefreshToken{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RefreshToken{}, fmt.Errorf("commit refresh token rotation: %w", err)
	}

	return createdRefreshToken, nil
}

func findRefreshTokenForUpdate(ctx context.Context, tx pgx.Tx, tokenHash string) (RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, replaced_by_token_id, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`

	var refreshToken RefreshToken
	err := tx.QueryRow(ctx, query, tokenHash).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.TokenHash,
		&refreshToken.ExpiresAt,
		&refreshToken.RevokedAt,
		&refreshToken.ReplacedByTokenID,
		&refreshToken.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefreshToken{}, ErrInvalidRefreshToken
	}
	if err != nil {
		return RefreshToken{}, fmt.Errorf("cannot find refresh token: %w", err)
	}

	return refreshToken, nil
}

func createRefreshToken(ctx context.Context, tx pgx.Tx, refreshToken RefreshToken) (RefreshToken, error) {
	query := `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, replaced_by_token_id, created_at
	`

	var createdRefreshToken RefreshToken
	err := tx.QueryRow(
		ctx,
		query,
		refreshToken.ID,
		refreshToken.UserID,
		refreshToken.TokenHash,
		refreshToken.ExpiresAt,
	).Scan(
		&createdRefreshToken.ID,
		&createdRefreshToken.UserID,
		&createdRefreshToken.TokenHash,
		&createdRefreshToken.ExpiresAt,
		&createdRefreshToken.RevokedAt,
		&createdRefreshToken.ReplacedByTokenID,
		&createdRefreshToken.CreatedAt,
	)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("cannot create refresh token: %w", err)
	}

	return createdRefreshToken, nil
}

func revokeRefreshToken(
	ctx context.Context,
	tx pgx.Tx,
	refreshToken RefreshToken,
	revokedAt time.Time,
	replacedByTokenID *uuid.UUID,
) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, $2),
			replaced_by_token_id = $3
		WHERE id = $1
	`

	if _, err := tx.Exec(ctx, query, refreshToken.ID, revokedAt, replacedByTokenID); err != nil {
		return fmt.Errorf("cannot revoke refresh token: %w", err)
	}

	return nil
}
