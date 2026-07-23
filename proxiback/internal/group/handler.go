package group

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/auth"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

type createGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type createInvitationRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
	MaxUses   *int       `json:"max_uses"`
}

type joinGroupRequest struct {
	InvitationCode      string `json:"invitationCode"`
	InvitationCodeSnake string `json:"invitation_code"`
}

type updateMemberRequest struct {
	Role MemberRole `json:"role"`
}

type updateInvitationRequest struct {
	Active *bool `json:"active"`
}

type groupResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	OwnerID     string      `json:"owner_id"`
	Role        MemberRole  `json:"role"`
	Permissions Permissions `json:"permissions"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type memberResponse struct {
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Role        MemberRole `json:"role"`
	JoinedAt    time.Time  `json:"joined_at"`
}

type updatedMemberResponse struct {
	UserID   string     `json:"user_id"`
	Role     MemberRole `json:"role"`
	JoinedAt time.Time  `json:"joined_at"`
}

type invitationResponse struct {
	ID            string     `json:"id"`
	GroupID       string     `json:"group_id"`
	Code          string     `json:"code"`
	ExpiresAt     time.Time  `json:"expires_at"`
	MaxUses       int        `json:"max_uses"`
	UsesCount     int        `json:"uses_count"`
	Active        bool       `json:"active"`
	CreatedBy     string     `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (handler *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	var request createGroupRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	membership, err := handler.service.CreateGroup(r.Context(), userID, request.Name, request.Description)
	if err != nil {
		handler.writeServiceError(w, err, "failed to create group")
		return
	}

	writeJSON(w, http.StatusCreated, newGroupResponse(membership))
}

func (handler *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	memberships, err := handler.service.ListGroups(r.Context(), userID)
	if err != nil {
		handler.writeServiceError(w, err, "failed to list groups")
		return
	}

	responses := make([]groupResponse, 0, len(memberships))
	for _, membership := range memberships {
		responses = append(responses, newGroupResponse(membership))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (handler *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	groupID, ok := parsePathUUID(w, r, "groupId", "INVALID_GROUP_ID", "L'identifiant du groupe est invalide.")
	if !ok {
		return
	}

	membership, err := handler.service.GetGroup(r.Context(), userID, groupID)
	if err != nil {
		handler.writeServiceError(w, err, "failed to get group")
		return
	}

	writeJSON(w, http.StatusOK, newGroupResponse(membership))
}

func (handler *Handler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	groupID, ok := parsePathUUID(w, r, "groupId", "INVALID_GROUP_ID", "L'identifiant du groupe est invalide.")
	if !ok {
		return
	}

	var request createInvitationRequest
	if !decodeOptionalJSON(w, r, &request) {
		return
	}

	invitation, err := handler.service.CreateInvitation(r.Context(), userID, groupID, CreateInvitationInput{
		ExpiresAt: request.ExpiresAt,
		MaxUses:   request.MaxUses,
	})
	if err != nil {
		handler.writeServiceError(w, err, "failed to create group invitation")
		return
	}

	writeJSON(w, http.StatusCreated, newInvitationResponse(invitation))
}

func (handler *Handler) DeactivateInvitation(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	groupID, ok := parsePathUUID(w, r, "groupId", "INVALID_GROUP_ID", "L'identifiant du groupe est invalide.")
	if !ok {
		return
	}

	invitationID, ok := parsePathUUID(w, r, "invitationId", "INVALID_INVITATION_ID", "L'identifiant de l'invitation est invalide.")
	if !ok {
		return
	}

	var request updateInvitationRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	if request.Active == nil || *request.Active {
		handler.writeServiceError(w, ErrInvalidInvitationUpdate, "invalid invitation update")
		return
	}

	invitation, err := handler.service.DeactivateInvitation(r.Context(), userID, groupID, invitationID)
	if err != nil {
		handler.writeServiceError(w, err, "failed to deactivate group invitation")
		return
	}

	writeJSON(w, http.StatusOK, newInvitationResponse(invitation))
}

func (handler *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	var request joinGroupRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	invitationCode := request.InvitationCode
	if invitationCode == "" {
		invitationCode = request.InvitationCodeSnake
	}

	membership, err := handler.service.JoinGroup(r.Context(), userID, invitationCode)
	if err != nil {
		handler.writeServiceError(w, err, "failed to join group")
		return
	}

	writeJSON(w, http.StatusCreated, newGroupResponse(membership))
}

func (handler *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	groupID, ok := parsePathUUID(w, r, "groupId", "INVALID_GROUP_ID", "L'identifiant du groupe est invalide.")
	if !ok {
		return
	}

	members, err := handler.service.ListMembers(r.Context(), userID, groupID)
	if err != nil {
		handler.writeServiceError(w, err, "failed to list group members")
		return
	}

	responses := make([]memberResponse, 0, len(members))
	for _, member := range members {
		responses = append(responses, newMemberResponse(member))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (handler *Handler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	groupID, ok := parsePathUUID(w, r, "groupId", "INVALID_GROUP_ID", "L'identifiant du groupe est invalide.")
	if !ok {
		return
	}

	targetUserID, ok := parsePathUUID(w, r, "userId", "INVALID_USER_ID", "L'identifiant de l'utilisateur est invalide.")
	if !ok {
		return
	}

	var request updateMemberRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	member, err := handler.service.UpdateMemberRole(r.Context(), userID, groupID, targetUserID, request.Role)
	if err != nil {
		handler.writeServiceError(w, err, "failed to update group member")
		return
	}

	writeJSON(w, http.StatusOK, updatedMemberResponse{
		UserID:   member.UserID.String(),
		Role:     member.Role,
		JoinedAt: member.JoinedAt,
	})
}

func (handler *Handler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	groupID, ok := parsePathUUID(w, r, "groupId", "INVALID_GROUP_ID", "L'identifiant du groupe est invalide.")
	if !ok {
		return
	}

	targetUserID, ok := parsePathUUID(w, r, "userId", "INVALID_USER_ID", "L'identifiant de l'utilisateur est invalide.")
	if !ok {
		return
	}

	if err := handler.service.DeleteMember(r.Context(), userID, groupID, targetUserID); err != nil {
		handler.writeServiceError(w, err, "failed to delete group member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	groupID, ok := parsePathUUID(w, r, "groupId", "INVALID_GROUP_ID", "L'identifiant du groupe est invalide.")
	if !ok {
		return
	}

	if err := handler.service.DeleteGroup(r.Context(), userID, groupID); err != nil {
		handler.writeServiceError(w, err, "failed to delete group")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newGroupResponse(membership Membership) groupResponse {
	return groupResponse{
		ID:          membership.Group.ID.String(),
		Name:        membership.Group.Name,
		Description: membership.Group.Description,
		OwnerID:     membership.Group.OwnerID.String(),
		Role:        membership.Role,
		Permissions: membership.Role.Permissions(),
		CreatedAt:   membership.Group.CreatedAt,
		UpdatedAt:   membership.Group.UpdatedAt,
	}
}

func newMemberResponse(member MemberWithUser) memberResponse {
	return memberResponse{
		UserID:      member.UserID.String(),
		DisplayName: member.DisplayName,
		Email:       member.Email,
		Role:        member.Role,
		JoinedAt:    member.JoinedAt,
	}
}

func newInvitationResponse(invitation Invitation) invitationResponse {
	return invitationResponse{
		ID:            invitation.ID.String(),
		GroupID:       invitation.GroupID.String(),
		Code:          invitation.Code,
		ExpiresAt:     invitation.ExpiresAt,
		MaxUses:       invitation.MaxUses,
		UsesCount:     invitation.UsesCount,
		Active:        invitation.Active,
		CreatedBy:     invitation.CreatedBy.String(),
		CreatedAt:     invitation.CreatedAt,
		UpdatedAt:     invitation.UpdatedAt,
		DeactivatedAt: invitation.DeactivatedAt,
	}
}

func authenticatedUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Le token est invalide.")
		return uuid.Nil, false
	}

	return userID, true
}

func parsePathUUID(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	code string,
	message string,
) (uuid.UUID, bool) {
	value, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeError(w, http.StatusBadRequest, code, message)
		return uuid.Nil, false
	}

	return value, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(value); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Le corps de la requete est invalide.")
		return false
	}

	return true
}

func decodeOptionalJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(value); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}

		writeError(w, http.StatusBadRequest, "invalid_request", "Le corps de la requete est invalide.")
		return false
	}

	return true
}

func (handler *Handler) writeServiceError(w http.ResponseWriter, err error, logMessage string) {
	switch {
	case errors.Is(err, ErrInvalidGroupName):
		writeError(w, http.StatusBadRequest, "INVALID_GROUP_NAME", "Le nom du groupe doit contenir entre 2 et 100 caracteres.")

	case errors.Is(err, ErrInvalidGroupDescription):
		writeError(w, http.StatusBadRequest, "INVALID_GROUP_DESCRIPTION", "La description du groupe ne doit pas depasser 500 caracteres.")

	case errors.Is(err, ErrInvalidInvitationExpiration):
		writeError(w, http.StatusBadRequest, "INVALID_INVITATION_EXPIRATION", "La date d'expiration doit etre dans le futur.")

	case errors.Is(err, ErrInvalidInvitationMaxUses):
		writeError(w, http.StatusBadRequest, "INVALID_INVITATION_MAX_USES", "Le nombre maximal d'utilisations doit etre compris entre 1 et 1000.")

	case errors.Is(err, ErrInvalidInvitationCode):
		writeError(w, http.StatusBadRequest, "INVALID_INVITATION_CODE", "Le code d'invitation est invalide.")

	case errors.Is(err, ErrInvalidMemberRole):
		writeError(w, http.StatusBadRequest, "INVALID_MEMBER_ROLE", "Le role doit etre OWNER, ADMIN ou MEMBER.")

	case errors.Is(err, ErrInvalidInvitationUpdate):
		writeError(w, http.StatusBadRequest, "INVALID_INVITATION_UPDATE", "Seule la desactivation du code est autorisee.")

	case errors.Is(err, ErrGroupNotFound):
		writeError(w, http.StatusNotFound, "GROUP_NOT_FOUND", "Le groupe est introuvable.")

	case errors.Is(err, ErrMemberNotFound):
		writeError(w, http.StatusNotFound, "MEMBER_NOT_FOUND", "Le membre est introuvable.")

	case errors.Is(err, ErrInvitationNotFound):
		writeError(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "Le code d'invitation est introuvable.")

	case errors.Is(err, ErrInvitationInactive):
		writeError(w, http.StatusGone, "INVITATION_INACTIVE", "Le code d'invitation a ete desactive.")

	case errors.Is(err, ErrInvitationExpired):
		writeError(w, http.StatusGone, "INVITATION_EXPIRED", "Le code d'invitation a expire.")

	case errors.Is(err, ErrInvitationMaxUsesReached):
		writeError(w, http.StatusGone, "INVITATION_MAX_USES_REACHED", "Le code d'invitation a atteint son nombre maximal d'utilisations.")

	case errors.Is(err, ErrAlreadyMember):
		writeError(w, http.StatusConflict, "ALREADY_MEMBER", "L'utilisateur est deja membre du groupe.")

	case errors.Is(err, ErrForbidden), errors.Is(err, ErrCannotChangeOwnerRole), errors.Is(err, ErrCannotRemoveOwner):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Vous n'avez pas les droits necessaires pour cette action.")

	default:
		handler.logger.Error(logMessage, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}

func writeJSON(responseWriter http.ResponseWriter, status int, value any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(status)

	_ = json.NewEncoder(responseWriter).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}
