package group

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestServiceCreateGroupMakesCreatorOwner(t *testing.T) {
	t.Parallel()

	repository := newFakeRepository()
	service := NewService(repository)
	userID := uuid.New()

	membership, err := service.CreateGroup(context.Background(), userID, " Proxiad Lille ", " Les paris de l'agence ")
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if membership.Group.Name != "Proxiad Lille" {
		t.Fatalf("group name = %q, want %q", membership.Group.Name, "Proxiad Lille")
	}

	if membership.Group.OwnerID != userID {
		t.Fatalf("ownerID = %s, want %s", membership.Group.OwnerID, userID)
	}

	if membership.Role != RoleOwner {
		t.Fatalf("role = %s, want %s", membership.Role, RoleOwner)
	}

	permissions := membership.Role.Permissions()
	if !permissions.CanCreateBet || !permissions.CanDeclareResult || !permissions.CanDeleteGroup {
		t.Fatalf("owner permissions = %+v, want admin and delete permissions", permissions)
	}
}

func TestServiceCreateInvitationUsesDefaultsAndRetriesDuplicateCode(t *testing.T) {
	t.Parallel()

	repository := newFakeRepository()
	service := NewService(repository)
	now := time.Date(2026, time.July, 23, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	codes := []string{"PROXI-LILLE-7FK2", "PROXI-LILLE-9Q2M"}
	service.invitationCodeGenerator = func(string) (string, error) {
		code := codes[0]
		codes = codes[1:]
		return code, nil
	}

	ownerID := uuid.New()
	groupID := repository.addGroup("Proxiad Lille", ownerID)
	repository.createInvitationErrors = []error{ErrInvitationCodeAlreadyExists}

	invitation, err := service.CreateInvitation(context.Background(), ownerID, groupID, CreateInvitationInput{})
	if err != nil {
		t.Fatalf("CreateInvitation() error = %v", err)
	}

	if invitation.Code != "PROXI-LILLE-9Q2M" {
		t.Fatalf("code = %q, want retry code", invitation.Code)
	}

	if !invitation.ExpiresAt.Equal(now.Add(defaultInvitationExpiration)) {
		t.Fatalf("expiresAt = %s, want %s", invitation.ExpiresAt, now.Add(defaultInvitationExpiration))
	}

	if invitation.MaxUses != defaultInvitationMaxUses {
		t.Fatalf("maxUses = %d, want %d", invitation.MaxUses, defaultInvitationMaxUses)
	}
}

func TestServiceCreateInvitationRejectsMemberRole(t *testing.T) {
	t.Parallel()

	repository := newFakeRepository()
	service := NewService(repository)
	memberID := uuid.New()
	groupID := repository.addGroup("Proxiad Lille", uuid.New())
	repository.members[memberKey(groupID, memberID)] = Member{GroupID: groupID, UserID: memberID, Role: RoleMember}

	_, err := service.CreateInvitation(context.Background(), memberID, groupID, CreateInvitationInput{})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("CreateInvitation() error = %v, want %v", err, ErrForbidden)
	}
}

func TestServiceCreateInvitationRejectsExplicitZeroMaxUses(t *testing.T) {
	t.Parallel()

	repository := newFakeRepository()
	service := NewService(repository)
	ownerID := uuid.New()
	groupID := repository.addGroup("Proxiad Lille", ownerID)
	maxUses := 0

	_, err := service.CreateInvitation(context.Background(), ownerID, groupID, CreateInvitationInput{MaxUses: &maxUses})
	if !errors.Is(err, ErrInvalidInvitationMaxUses) {
		t.Fatalf("CreateInvitation() error = %v, want %v", err, ErrInvalidInvitationMaxUses)
	}
}

func TestServiceUpdateMemberRoleTransfersOwnershipFromOwner(t *testing.T) {
	t.Parallel()

	repository := newFakeRepository()
	service := NewService(repository)
	ownerID := uuid.New()
	nextOwnerID := uuid.New()
	groupID := repository.addGroup("Proxiad Lille", ownerID)
	repository.members[memberKey(groupID, nextOwnerID)] = Member{GroupID: groupID, UserID: nextOwnerID, Role: RoleMember}

	member, err := service.UpdateMemberRole(context.Background(), ownerID, groupID, nextOwnerID, RoleOwner)
	if err != nil {
		t.Fatalf("UpdateMemberRole() error = %v", err)
	}

	if member.Role != RoleOwner {
		t.Fatalf("next owner role = %s, want %s", member.Role, RoleOwner)
	}

	previousOwner := repository.members[memberKey(groupID, ownerID)]
	if previousOwner.Role != RoleAdmin {
		t.Fatalf("previous owner role = %s, want %s", previousOwner.Role, RoleAdmin)
	}

	if repository.groups[groupID].OwnerID != nextOwnerID {
		t.Fatalf("group ownerID = %s, want %s", repository.groups[groupID].OwnerID, nextOwnerID)
	}
}

func TestServiceUpdateMemberRoleRejectsOwnerPromotionByAdmin(t *testing.T) {
	t.Parallel()

	repository := newFakeRepository()
	service := NewService(repository)
	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	groupID := repository.addGroup("Proxiad Lille", ownerID)
	repository.members[memberKey(groupID, adminID)] = Member{GroupID: groupID, UserID: adminID, Role: RoleAdmin}
	repository.members[memberKey(groupID, memberID)] = Member{GroupID: groupID, UserID: memberID, Role: RoleMember}

	_, err := service.UpdateMemberRole(context.Background(), adminID, groupID, memberID, RoleOwner)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("UpdateMemberRole() error = %v, want %v", err, ErrForbidden)
	}
}

func TestMemberRolePermissions(t *testing.T) {
	t.Parallel()

	memberPermissions := RoleMember.Permissions()
	if !memberPermissions.CanViewBets || !memberPermissions.CanPlaceBet {
		t.Fatalf("member permissions = %+v, want view and place bet", memberPermissions)
	}

	if memberPermissions.CanCreateBet || memberPermissions.CanCancelBet || memberPermissions.CanDeclareResult || memberPermissions.CanDeleteGroup {
		t.Fatalf("member permissions = %+v, want no admin permissions", memberPermissions)
	}

	adminPermissions := RoleAdmin.Permissions()
	if !adminPermissions.CanCreateBet || !adminPermissions.CanCancelBet || !adminPermissions.CanDeclareResult {
		t.Fatalf("admin permissions = %+v, want bet administration permissions", adminPermissions)
	}

	if adminPermissions.CanDeleteGroup {
		t.Fatalf("admin permissions = %+v, want no group delete permission", adminPermissions)
	}
}

func TestInvitationCodeBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		groupName string
		want      string
	}{
		{name: "proxiad agency", groupName: "Proxiad Lille", want: "PROXI-LILLE"},
		{name: "private circle", groupName: "Poker Friday", want: "PROXI-POKER-FRIDAY"},
		{name: "blank fallback", groupName: "---", want: "PROXI"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := invitationCodeBase(test.groupName); got != test.want {
				t.Fatalf("invitationCodeBase(%q) = %q, want %q", test.groupName, got, test.want)
			}
		})
	}
}

type fakeRepository struct {
	groups                 map[uuid.UUID]Group
	members                map[string]Member
	invitations            map[uuid.UUID]Invitation
	createInvitationErrors []error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		groups:      map[uuid.UUID]Group{},
		members:     map[string]Member{},
		invitations: map[uuid.UUID]Invitation{},
	}
}

func (repository *fakeRepository) addGroup(name string, ownerID uuid.UUID) uuid.UUID {
	groupID := uuid.New()
	repository.groups[groupID] = Group{ID: groupID, Name: name, OwnerID: ownerID}
	repository.members[memberKey(groupID, ownerID)] = Member{GroupID: groupID, UserID: ownerID, Role: RoleOwner}
	return groupID
}

func (repository *fakeRepository) CreateGroup(_ context.Context, newGroup Group, ownerID uuid.UUID) (Membership, error) {
	repository.groups[newGroup.ID] = newGroup
	repository.members[memberKey(newGroup.ID, ownerID)] = Member{GroupID: newGroup.ID, UserID: ownerID, Role: RoleOwner}
	return Membership{Group: newGroup, Role: RoleOwner}, nil
}

func (repository *fakeRepository) ListGroupsForUser(_ context.Context, userID uuid.UUID) ([]Membership, error) {
	memberships := make([]Membership, 0)
	for _, member := range repository.members {
		if member.UserID != userID {
			continue
		}

		foundGroup, ok := repository.groups[member.GroupID]
		if ok {
			memberships = append(memberships, Membership{Group: foundGroup, Role: member.Role})
		}
	}

	return memberships, nil
}

func (repository *fakeRepository) FindGroupForUser(_ context.Context, groupID uuid.UUID, userID uuid.UUID) (Membership, error) {
	member, ok := repository.members[memberKey(groupID, userID)]
	if !ok {
		return Membership{}, ErrGroupNotFound
	}

	foundGroup, ok := repository.groups[groupID]
	if !ok {
		return Membership{}, ErrGroupNotFound
	}

	return Membership{Group: foundGroup, Role: member.Role}, nil
}

func (repository *fakeRepository) FindMember(_ context.Context, groupID uuid.UUID, userID uuid.UUID) (Member, error) {
	member, ok := repository.members[memberKey(groupID, userID)]
	if !ok {
		return Member{}, ErrMemberNotFound
	}

	return member, nil
}

func (repository *fakeRepository) ListMembers(_ context.Context, groupID uuid.UUID) ([]MemberWithUser, error) {
	members := make([]MemberWithUser, 0)
	for _, member := range repository.members {
		if member.GroupID == groupID {
			members = append(members, MemberWithUser{Member: member})
		}
	}

	return members, nil
}

func (repository *fakeRepository) UpdateMemberRole(_ context.Context, groupID uuid.UUID, userID uuid.UUID, role MemberRole) (Member, error) {
	key := memberKey(groupID, userID)
	member, ok := repository.members[key]
	if !ok {
		return Member{}, ErrMemberNotFound
	}

	member.Role = role
	repository.members[key] = member
	return member, nil
}

func (repository *fakeRepository) TransferOwnership(_ context.Context, groupID uuid.UUID, nextOwnerID uuid.UUID) (Member, error) {
	for key, member := range repository.members {
		if member.GroupID == groupID && member.Role == RoleOwner {
			member.Role = RoleAdmin
			repository.members[key] = member
		}
	}

	key := memberKey(groupID, nextOwnerID)
	nextOwner, ok := repository.members[key]
	if !ok {
		return Member{}, ErrMemberNotFound
	}

	nextOwner.Role = RoleOwner
	repository.members[key] = nextOwner

	foundGroup := repository.groups[groupID]
	foundGroup.OwnerID = nextOwnerID
	repository.groups[groupID] = foundGroup

	return nextOwner, nil
}

func (repository *fakeRepository) DeleteMember(_ context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	key := memberKey(groupID, userID)
	if _, ok := repository.members[key]; !ok {
		return ErrMemberNotFound
	}

	delete(repository.members, key)
	return nil
}

func (repository *fakeRepository) DeleteGroup(_ context.Context, groupID uuid.UUID) error {
	if _, ok := repository.groups[groupID]; !ok {
		return ErrGroupNotFound
	}

	delete(repository.groups, groupID)
	return nil
}

func (repository *fakeRepository) CreateInvitation(_ context.Context, invitation Invitation) (Invitation, error) {
	if len(repository.createInvitationErrors) > 0 {
		err := repository.createInvitationErrors[0]
		repository.createInvitationErrors = repository.createInvitationErrors[1:]
		return Invitation{}, err
	}

	repository.invitations[invitation.ID] = invitation
	return invitation, nil
}

func (repository *fakeRepository) DeactivateInvitation(_ context.Context, groupID uuid.UUID, invitationID uuid.UUID) (Invitation, error) {
	invitation, ok := repository.invitations[invitationID]
	if !ok || invitation.GroupID != groupID {
		return Invitation{}, ErrInvitationNotFound
	}

	invitation.Active = false
	repository.invitations[invitationID] = invitation
	return invitation, nil
}

func (repository *fakeRepository) JoinGroupByInvitation(_ context.Context, userID uuid.UUID, invitationCode string, now time.Time) (Membership, error) {
	for _, invitation := range repository.invitations {
		if invitation.Code != invitationCode {
			continue
		}

		if !invitation.Active {
			return Membership{}, ErrInvitationInactive
		}

		if !now.Before(invitation.ExpiresAt) {
			return Membership{}, ErrInvitationExpired
		}

		key := memberKey(invitation.GroupID, userID)
		if _, ok := repository.members[key]; ok {
			return Membership{}, ErrAlreadyMember
		}

		repository.members[key] = Member{GroupID: invitation.GroupID, UserID: userID, Role: RoleMember}
		return Membership{Group: repository.groups[invitation.GroupID], Role: RoleMember}, nil
	}

	return Membership{}, ErrInvitationNotFound
}

func memberKey(groupID uuid.UUID, userID uuid.UUID) string {
	return groupID.String() + ":" + userID.String()
}
