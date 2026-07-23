package group

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

func (repository *PostgresRepository) CreateGroup(
	ctx context.Context,
	newGroup Group,
	ownerID uuid.UUID,
) (Membership, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Membership{}, fmt.Errorf("cannot start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO groups (id, name, description, owner_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, owner_id, created_at, updated_at
	`

	var createdGroup Group
	err = tx.QueryRow(
		ctx,
		query,
		newGroup.ID,
		newGroup.Name,
		newGroup.Description,
		ownerID,
	).Scan(
		&createdGroup.ID,
		&createdGroup.Name,
		&createdGroup.Description,
		&createdGroup.OwnerID,
		&createdGroup.CreatedAt,
		&createdGroup.UpdatedAt,
	)
	if err != nil {
		return Membership{}, fmt.Errorf("cannot create group: %w", err)
	}

	memberQuery := `
		INSERT INTO group_members (group_id, user_id, role)
		VALUES ($1, $2, $3)
	`
	if _, err := tx.Exec(ctx, memberQuery, createdGroup.ID, ownerID, RoleOwner); err != nil {
		return Membership{}, fmt.Errorf("cannot create group owner membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Membership{}, fmt.Errorf("cannot commit group creation: %w", err)
	}

	return Membership{
		Group: createdGroup,
		Role:  RoleOwner,
	}, nil
}

func (repository *PostgresRepository) ListGroupsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]Membership, error) {
	query := `
		SELECT g.id, g.name, g.description, g.owner_id, g.created_at, g.updated_at, m.role
		FROM groups g
		INNER JOIN group_members m ON m.group_id = g.id
		WHERE m.user_id = $1
		ORDER BY g.created_at DESC
	`

	rows, err := repository.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot list groups: %w", err)
	}
	defer rows.Close()

	memberships := make([]Membership, 0)
	for rows.Next() {
		var membership Membership
		if err := rows.Scan(
			&membership.Group.ID,
			&membership.Group.Name,
			&membership.Group.Description,
			&membership.Group.OwnerID,
			&membership.Group.CreatedAt,
			&membership.Group.UpdatedAt,
			&membership.Role,
		); err != nil {
			return nil, fmt.Errorf("cannot scan group: %w", err)
		}

		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate groups: %w", err)
	}

	return memberships, nil
}

func (repository *PostgresRepository) FindGroupForUser(
	ctx context.Context,
	groupID uuid.UUID,
	userID uuid.UUID,
) (Membership, error) {
	query := `
		SELECT g.id, g.name, g.description, g.owner_id, g.created_at, g.updated_at, m.role
		FROM groups g
		INNER JOIN group_members m ON m.group_id = g.id
		WHERE g.id = $1 AND m.user_id = $2
	`

	var membership Membership
	err := repository.pool.QueryRow(ctx, query, groupID, userID).Scan(
		&membership.Group.ID,
		&membership.Group.Name,
		&membership.Group.Description,
		&membership.Group.OwnerID,
		&membership.Group.CreatedAt,
		&membership.Group.UpdatedAt,
		&membership.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Membership{}, ErrGroupNotFound
		}

		return Membership{}, fmt.Errorf("cannot find group: %w", err)
	}

	return membership, nil
}

func (repository *PostgresRepository) FindMember(
	ctx context.Context,
	groupID uuid.UUID,
	userID uuid.UUID,
) (Member, error) {
	query := `
		SELECT group_id, user_id, role, joined_at, updated_at
		FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`

	var member Member
	err := repository.pool.QueryRow(ctx, query, groupID, userID).Scan(
		&member.GroupID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&member.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrMemberNotFound
		}

		return Member{}, fmt.Errorf("cannot find member: %w", err)
	}

	return member, nil
}

func (repository *PostgresRepository) ListMembers(
	ctx context.Context,
	groupID uuid.UUID,
) ([]MemberWithUser, error) {
	query := `
		SELECT m.group_id, m.user_id, m.role, m.joined_at, m.updated_at, u.display_name, u.email
		FROM group_members m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.group_id = $1
		ORDER BY
			CASE m.role
				WHEN 'OWNER' THEN 1
				WHEN 'ADMIN' THEN 2
				ELSE 3
			END,
			u.display_name ASC
	`

	rows, err := repository.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("cannot list group members: %w", err)
	}
	defer rows.Close()

	members := make([]MemberWithUser, 0)
	for rows.Next() {
		var member MemberWithUser
		if err := rows.Scan(
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.UpdatedAt,
			&member.DisplayName,
			&member.Email,
		); err != nil {
			return nil, fmt.Errorf("cannot scan group member: %w", err)
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate group members: %w", err)
	}

	return members, nil
}

func (repository *PostgresRepository) UpdateMemberRole(
	ctx context.Context,
	groupID uuid.UUID,
	userID uuid.UUID,
	role MemberRole,
) (Member, error) {
	query := `
		UPDATE group_members
		SET role = $3,
			updated_at = NOW()
		WHERE group_id = $1 AND user_id = $2
		RETURNING group_id, user_id, role, joined_at, updated_at
	`

	var updatedMember Member
	err := repository.pool.QueryRow(ctx, query, groupID, userID, role).Scan(
		&updatedMember.GroupID,
		&updatedMember.UserID,
		&updatedMember.Role,
		&updatedMember.JoinedAt,
		&updatedMember.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrMemberNotFound
		}

		return Member{}, fmt.Errorf("cannot update group member: %w", err)
	}

	return updatedMember, nil
}

func (repository *PostgresRepository) TransferOwnership(
	ctx context.Context,
	groupID uuid.UUID,
	nextOwnerID uuid.UUID,
) (Member, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Member{}, fmt.Errorf("cannot start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(
		ctx,
		`
			UPDATE group_members
			SET role = $2,
				updated_at = NOW()
			WHERE group_id = $1 AND role = $3
		`,
		groupID,
		RoleAdmin,
		RoleOwner,
	); err != nil {
		return Member{}, fmt.Errorf("cannot demote previous owner: %w", err)
	}

	var nextOwner Member
	err = tx.QueryRow(
		ctx,
		`
			UPDATE group_members
			SET role = $3,
				updated_at = NOW()
			WHERE group_id = $1 AND user_id = $2
			RETURNING group_id, user_id, role, joined_at, updated_at
		`,
		groupID,
		nextOwnerID,
		RoleOwner,
	).Scan(
		&nextOwner.GroupID,
		&nextOwner.UserID,
		&nextOwner.Role,
		&nextOwner.JoinedAt,
		&nextOwner.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrMemberNotFound
		}

		return Member{}, fmt.Errorf("cannot promote next owner: %w", err)
	}

	commandTag, err := tx.Exec(
		ctx,
		`
			UPDATE groups
			SET owner_id = $2,
				updated_at = NOW()
			WHERE id = $1
		`,
		groupID,
		nextOwnerID,
	)
	if err != nil {
		return Member{}, fmt.Errorf("cannot update group owner: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return Member{}, ErrGroupNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return Member{}, fmt.Errorf("cannot commit ownership transfer: %w", err)
	}

	return nextOwner, nil
}

func (repository *PostgresRepository) DeleteMember(
	ctx context.Context,
	groupID uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := repository.pool.Exec(
		ctx,
		`
			DELETE FROM group_members
			WHERE group_id = $1 AND user_id = $2
		`,
		groupID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("cannot delete group member: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (repository *PostgresRepository) DeleteGroup(ctx context.Context, groupID uuid.UUID) error {
	commandTag, err := repository.pool.Exec(
		ctx,
		`
			DELETE FROM groups
			WHERE id = $1
		`,
		groupID,
	)
	if err != nil {
		return fmt.Errorf("cannot delete group: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrGroupNotFound
	}

	return nil
}

func (repository *PostgresRepository) CreateInvitation(
	ctx context.Context,
	invitation Invitation,
) (Invitation, error) {
	query := `
		INSERT INTO group_invitations (
			id,
			group_id,
			code,
			expires_at,
			max_uses,
			active,
			created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, group_id, code, expires_at, max_uses, uses_count, active, created_by, created_at, updated_at, deactivated_at
	`

	createdInvitation, err := scanInvitation(repository.pool.QueryRow(
		ctx,
		query,
		invitation.ID,
		invitation.GroupID,
		invitation.Code,
		invitation.ExpiresAt,
		invitation.MaxUses,
		invitation.Active,
		invitation.CreatedBy,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return Invitation{}, ErrInvitationCodeAlreadyExists
		}

		return Invitation{}, fmt.Errorf("cannot create group invitation: %w", err)
	}

	return createdInvitation, nil
}

func (repository *PostgresRepository) DeactivateInvitation(
	ctx context.Context,
	groupID uuid.UUID,
	invitationID uuid.UUID,
) (Invitation, error) {
	query := `
		UPDATE group_invitations
		SET active = FALSE,
			deactivated_at = COALESCE(deactivated_at, NOW()),
			updated_at = NOW()
		WHERE group_id = $1 AND id = $2
		RETURNING id, group_id, code, expires_at, max_uses, uses_count, active, created_by, created_at, updated_at, deactivated_at
	`

	invitation, err := scanInvitation(repository.pool.QueryRow(ctx, query, groupID, invitationID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invitation{}, ErrInvitationNotFound
		}

		return Invitation{}, fmt.Errorf("cannot deactivate group invitation: %w", err)
	}

	return invitation, nil
}

func (repository *PostgresRepository) JoinGroupByInvitation(
	ctx context.Context,
	userID uuid.UUID,
	invitationCode string,
	now time.Time,
) (Membership, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Membership{}, fmt.Errorf("cannot start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	invitation, err := scanInvitation(tx.QueryRow(
		ctx,
		`
			SELECT id, group_id, code, expires_at, max_uses, uses_count, active, created_by, created_at, updated_at, deactivated_at
			FROM group_invitations
			WHERE code = $1
			FOR UPDATE
		`,
		invitationCode,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Membership{}, ErrInvitationNotFound
		}

		return Membership{}, fmt.Errorf("cannot find group invitation: %w", err)
	}

	if !invitation.Active || invitation.DeactivatedAt != nil {
		return Membership{}, ErrInvitationInactive
	}

	if !now.UTC().Before(invitation.ExpiresAt) {
		return Membership{}, ErrInvitationExpired
	}

	if invitation.UsesCount >= invitation.MaxUses {
		return Membership{}, ErrInvitationMaxUsesReached
	}

	var existingMember Member
	err = tx.QueryRow(
		ctx,
		`
			SELECT group_id, user_id, role, joined_at, updated_at
			FROM group_members
			WHERE group_id = $1 AND user_id = $2
		`,
		invitation.GroupID,
		userID,
	).Scan(
		&existingMember.GroupID,
		&existingMember.UserID,
		&existingMember.Role,
		&existingMember.JoinedAt,
		&existingMember.UpdatedAt,
	)
	if err == nil {
		return Membership{}, ErrAlreadyMember
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, fmt.Errorf("cannot check group membership: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`
			INSERT INTO group_members (group_id, user_id, role)
			VALUES ($1, $2, $3)
		`,
		invitation.GroupID,
		userID,
		RoleMember,
	); err != nil {
		return Membership{}, fmt.Errorf("cannot create group membership: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`
			UPDATE group_invitations
			SET uses_count = uses_count + 1,
				updated_at = NOW()
			WHERE id = $1
		`,
		invitation.ID,
	); err != nil {
		return Membership{}, fmt.Errorf("cannot increment invitation uses: %w", err)
	}

	var membership Membership
	err = tx.QueryRow(
		ctx,
		`
			SELECT id, name, description, owner_id, created_at, updated_at
			FROM groups
			WHERE id = $1
		`,
		invitation.GroupID,
	).Scan(
		&membership.Group.ID,
		&membership.Group.Name,
		&membership.Group.Description,
		&membership.Group.OwnerID,
		&membership.Group.CreatedAt,
		&membership.Group.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Membership{}, ErrGroupNotFound
		}

		return Membership{}, fmt.Errorf("cannot find joined group: %w", err)
	}

	membership.Role = RoleMember

	if err := tx.Commit(ctx); err != nil {
		return Membership{}, fmt.Errorf("cannot commit group join: %w", err)
	}

	return membership, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanInvitation(row rowScanner) (Invitation, error) {
	var invitation Invitation
	var deactivatedAt pgtype.Timestamptz

	err := row.Scan(
		&invitation.ID,
		&invitation.GroupID,
		&invitation.Code,
		&invitation.ExpiresAt,
		&invitation.MaxUses,
		&invitation.UsesCount,
		&invitation.Active,
		&invitation.CreatedBy,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&deactivatedAt,
	)
	if err != nil {
		return Invitation{}, err
	}

	if deactivatedAt.Valid {
		value := deactivatedAt.Time
		invitation.DeactivatedAt = &value
	}

	return invitation, nil
}
