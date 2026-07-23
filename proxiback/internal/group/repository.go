package group

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateInvitationInput struct {
	ExpiresAt *time.Time
	MaxUses   *int
}

type Repository interface {
	CreateGroup(ctx context.Context, newGroup Group, ownerID uuid.UUID) (Membership, error)
	ListGroupsForUser(ctx context.Context, userID uuid.UUID) ([]Membership, error)
	FindGroupForUser(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (Membership, error)
	FindMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (Member, error)
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]MemberWithUser, error)
	UpdateMemberRole(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, role MemberRole) (Member, error)
	TransferOwnership(ctx context.Context, groupID uuid.UUID, nextOwnerID uuid.UUID) (Member, error)
	DeleteMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error
	DeleteGroup(ctx context.Context, groupID uuid.UUID) error
	CreateInvitation(ctx context.Context, invitation Invitation) (Invitation, error)
	DeactivateInvitation(ctx context.Context, groupID uuid.UUID, invitationID uuid.UUID) (Invitation, error)
	JoinGroupByInvitation(ctx context.Context, userID uuid.UUID, invitationCode string, now time.Time) (Membership, error)
}
