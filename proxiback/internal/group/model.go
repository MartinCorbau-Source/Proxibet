package group

import (
	"time"

	"github.com/google/uuid"
)

type MemberRole string

const (
	RoleOwner  MemberRole = "OWNER"
	RoleAdmin  MemberRole = "ADMIN"
	RoleMember MemberRole = "MEMBER"
)

type Group struct {
	ID          uuid.UUID
	Name        string
	Description string
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Member struct {
	GroupID   uuid.UUID
	UserID    uuid.UUID
	Role      MemberRole
	JoinedAt  time.Time
	UpdatedAt time.Time
}

type MemberWithUser struct {
	Member
	DisplayName string
	Email       string
}

type Invitation struct {
	ID            uuid.UUID
	GroupID       uuid.UUID
	Code          string
	ExpiresAt     time.Time
	MaxUses       int
	UsesCount     int
	Active        bool
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeactivatedAt *time.Time
}

type Membership struct {
	Group Group
	Role  MemberRole
}

type Permissions struct {
	CanViewBets      bool `json:"can_view_bets"`
	CanPlaceBet      bool `json:"can_place_bet"`
	CanCreateBet     bool `json:"can_create_bet"`
	CanCancelBet     bool `json:"can_cancel_bet"`
	CanDeclareResult bool `json:"can_declare_result"`
	CanDeleteGroup   bool `json:"can_delete_group"`
}

func (role MemberRole) IsValid() bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

func (role MemberRole) HasAdminPrivileges() bool {
	return role == RoleOwner || role == RoleAdmin
}

func (role MemberRole) Permissions() Permissions {
	return Permissions{
		CanViewBets:      role.IsValid(),
		CanPlaceBet:      role.IsValid(),
		CanCreateBet:     role.HasAdminPrivileges(),
		CanCancelBet:     role.HasAdminPrivileges(),
		CanDeclareResult: role.HasAdminPrivileges(),
		CanDeleteGroup:   role == RoleOwner,
	}
}
