package group

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	minGroupNameLength          = 2
	maxGroupNameLength          = 100
	maxGroupDescriptionLength   = 500
	defaultInvitationExpiration = 7 * 24 * time.Hour
	defaultInvitationMaxUses    = 1
	maxInvitationMaxUses        = 1000
	invitationCodeRetries       = 5
)

type Service struct {
	repository              Repository
	now                     func() time.Time
	invitationCodeGenerator func(string) (string, error)
}

func NewService(repository Repository) *Service {
	return &Service{
		repository:              repository,
		now:                     func() time.Time { return time.Now().UTC() },
		invitationCodeGenerator: generateInvitationCode,
	}
}

func (service *Service) CreateGroup(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	description string,
) (Membership, error) {
	name = normalizeText(name)
	description = normalizeText(description)

	if err := validateGroup(name, description); err != nil {
		return Membership{}, err
	}

	newGroup := Group{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		OwnerID:     userID,
	}

	return service.repository.CreateGroup(ctx, newGroup, userID)
}

func (service *Service) ListGroups(ctx context.Context, userID uuid.UUID) ([]Membership, error) {
	return service.repository.ListGroupsForUser(ctx, userID)
}

func (service *Service) GetGroup(ctx context.Context, userID uuid.UUID, groupID uuid.UUID) (Membership, error) {
	return service.repository.FindGroupForUser(ctx, groupID, userID)
}

func (service *Service) CreateInvitation(
	ctx context.Context,
	userID uuid.UUID,
	groupID uuid.UUID,
	input CreateInvitationInput,
) (Invitation, error) {
	member, err := service.repository.FindMember(ctx, groupID, userID)
	if err != nil {
		return Invitation{}, err
	}

	if !member.Role.HasAdminPrivileges() {
		return Invitation{}, ErrForbidden
	}

	now := service.now()
	expiresAt := now.Add(defaultInvitationExpiration)
	if input.ExpiresAt != nil {
		expiresAt = input.ExpiresAt.UTC()
	}

	if !expiresAt.After(now) {
		return Invitation{}, ErrInvalidInvitationExpiration
	}

	maxUses := defaultInvitationMaxUses
	if input.MaxUses != nil {
		maxUses = *input.MaxUses
	}

	if maxUses < 1 || maxUses > maxInvitationMaxUses {
		return Invitation{}, ErrInvalidInvitationMaxUses
	}

	groupMembership, err := service.repository.FindGroupForUser(ctx, groupID, userID)
	if err != nil {
		return Invitation{}, err
	}

	var lastErr error
	for range invitationCodeRetries {
		code, err := service.invitationCodeGenerator(groupMembership.Group.Name)
		if err != nil {
			return Invitation{}, fmt.Errorf("cannot generate invitation code: %w", err)
		}

		invitation := Invitation{
			ID:        uuid.New(),
			GroupID:   groupID,
			Code:      code,
			ExpiresAt: expiresAt,
			MaxUses:   maxUses,
			Active:    true,
			CreatedBy: userID,
		}

		createdInvitation, err := service.repository.CreateInvitation(ctx, invitation)
		if err == nil {
			return createdInvitation, nil
		}

		if !errors.Is(err, ErrInvitationCodeAlreadyExists) {
			return Invitation{}, err
		}

		lastErr = err
	}

	return Invitation{}, lastErr
}

func (service *Service) DeactivateInvitation(
	ctx context.Context,
	userID uuid.UUID,
	groupID uuid.UUID,
	invitationID uuid.UUID,
) (Invitation, error) {
	member, err := service.repository.FindMember(ctx, groupID, userID)
	if err != nil {
		return Invitation{}, err
	}

	if !member.Role.HasAdminPrivileges() {
		return Invitation{}, ErrForbidden
	}

	return service.repository.DeactivateInvitation(ctx, groupID, invitationID)
}

func (service *Service) JoinGroup(
	ctx context.Context,
	userID uuid.UUID,
	invitationCode string,
) (Membership, error) {
	invitationCode = normalizeInvitationCode(invitationCode)
	if invitationCode == "" {
		return Membership{}, ErrInvalidInvitationCode
	}

	return service.repository.JoinGroupByInvitation(ctx, userID, invitationCode, service.now())
}

func (service *Service) ListMembers(ctx context.Context, userID uuid.UUID, groupID uuid.UUID) ([]MemberWithUser, error) {
	if _, err := service.repository.FindMember(ctx, groupID, userID); err != nil {
		return nil, err
	}

	return service.repository.ListMembers(ctx, groupID)
}

func (service *Service) UpdateMemberRole(
	ctx context.Context,
	userID uuid.UUID,
	groupID uuid.UUID,
	targetUserID uuid.UUID,
	nextRole MemberRole,
) (Member, error) {
	if !nextRole.IsValid() {
		return Member{}, ErrInvalidMemberRole
	}

	actor, err := service.repository.FindMember(ctx, groupID, userID)
	if err != nil {
		return Member{}, err
	}

	target, err := service.repository.FindMember(ctx, groupID, targetUserID)
	if err != nil {
		return Member{}, err
	}

	if target.Role == RoleOwner {
		if actor.Role != RoleOwner {
			return Member{}, ErrForbidden
		}

		if nextRole != RoleOwner {
			return Member{}, ErrCannotChangeOwnerRole
		}

		return target, nil
	}

	if nextRole == RoleOwner {
		if actor.Role != RoleOwner {
			return Member{}, ErrForbidden
		}

		return service.repository.TransferOwnership(ctx, groupID, targetUserID)
	}

	if !actor.Role.HasAdminPrivileges() {
		return Member{}, ErrForbidden
	}

	return service.repository.UpdateMemberRole(ctx, groupID, targetUserID, nextRole)
}

func (service *Service) DeleteMember(
	ctx context.Context,
	userID uuid.UUID,
	groupID uuid.UUID,
	targetUserID uuid.UUID,
) error {
	actor, err := service.repository.FindMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	target, err := service.repository.FindMember(ctx, groupID, targetUserID)
	if err != nil {
		return err
	}

	if target.Role == RoleOwner {
		return ErrCannotRemoveOwner
	}

	if actor.Role.HasAdminPrivileges() || userID == targetUserID {
		return service.repository.DeleteMember(ctx, groupID, targetUserID)
	}

	return ErrForbidden
}

func (service *Service) DeleteGroup(ctx context.Context, userID uuid.UUID, groupID uuid.UUID) error {
	member, err := service.repository.FindMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	if member.Role != RoleOwner {
		return ErrForbidden
	}

	return service.repository.DeleteGroup(ctx, groupID)
}

func validateGroup(name string, description string) error {
	nameLength := utf8.RuneCountInString(name)
	if nameLength < minGroupNameLength || nameLength > maxGroupNameLength {
		return ErrInvalidGroupName
	}

	if utf8.RuneCountInString(description) > maxGroupDescriptionLength {
		return ErrInvalidGroupDescription
	}

	return nil
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

func normalizeInvitationCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func generateInvitationCode(groupName string) (string, error) {
	suffix, err := randomSuffix(4)
	if err != nil {
		return "", err
	}

	return invitationCodeBase(groupName) + "-" + suffix, nil
}

func invitationCodeBase(groupName string) string {
	words := invitationWords(groupName)
	if len(words) == 0 {
		return "PROXI"
	}

	if strings.HasPrefix(words[0], "PROXI") {
		words[0] = "PROXI"
	} else {
		words = append([]string{"PROXI"}, words...)
	}

	code := strings.Join(words, "-")
	if len(code) <= 40 {
		return code
	}

	return strings.TrimSuffix(code[:40], "-")
}

func invitationWords(groupName string) []string {
	fields := strings.FieldsFunc(groupName, func(r rune) bool {
		return !isASCIIAlnum(r)
	})

	words := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.ToUpper(strings.TrimSpace(field))
		if field != "" {
			words = append(words, field)
		}
	}

	return words
}

func isASCIIAlnum(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

func randomSuffix(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	var builder strings.Builder
	builder.Grow(length)

	for range length {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}

		builder.WriteByte(alphabet[index.Int64()])
	}

	return builder.String(), nil
}
