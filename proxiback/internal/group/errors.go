package group

import "errors"

var (
	ErrGroupNotFound               = errors.New("group not found")
	ErrMemberNotFound              = errors.New("member not found")
	ErrForbidden                   = errors.New("forbidden")
	ErrInvalidGroupName            = errors.New("invalid group name")
	ErrInvalidGroupDescription     = errors.New("invalid group description")
	ErrInvalidInvitationExpiration = errors.New("invalid invitation expiration")
	ErrInvalidInvitationMaxUses    = errors.New("invalid invitation max uses")
	ErrInvalidInvitationCode       = errors.New("invalid invitation code")
	ErrInvitationNotFound          = errors.New("invitation not found")
	ErrInvitationInactive          = errors.New("invitation inactive")
	ErrInvitationExpired           = errors.New("invitation expired")
	ErrInvitationMaxUsesReached    = errors.New("invitation max uses reached")
	ErrInvitationCodeAlreadyExists = errors.New("invitation code already exists")
	ErrAlreadyMember               = errors.New("already member")
	ErrInvalidMemberRole           = errors.New("invalid member role")
	ErrCannotChangeOwnerRole       = errors.New("cannot change owner role")
	ErrCannotRemoveOwner           = errors.New("cannot remove owner")
	ErrInvalidInvitationUpdate     = errors.New("invalid invitation update")
)
