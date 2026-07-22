package profile

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"unicode/utf8"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/auth"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

type Handler struct {
	userRepo user.Repository
	logger   *slog.Logger
}

type Response struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type UpdateRequest struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email"`
}

func NewHandler(userRepo user.Repository, logger *slog.Logger) *Handler {
	return &Handler{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Le token est invalide.")
		return
	}

	foundUser, err := h.userRepo.FindByID(r.Context(), userID)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, newResponse(foundUser))

	case errors.Is(err, user.ErrUserNotFound):
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Le token est invalide.")

	default:
		h.logger.Error("failed to get profile", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Le token est invalide.")
		return
	}

	var updateRequest UpdateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&updateRequest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Le corps de la requete est invalide.")
		return
	}

	update, err := validateUpdateRequest(updateRequest)
	if err != nil {
		writeValidationError(w, err)
		return
	}

	updatedUser, err := h.userRepo.UpdateProfile(r.Context(), userID, update)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, newResponse(updatedUser))

	case errors.Is(err, user.ErrEmailAlreadyInUse):
		writeError(w, http.StatusConflict, "EMAIL_ALREADY_USED", "Cette adresse email est deja utilisee.")

	case errors.Is(err, user.ErrUserNotFound):
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Le token est invalide.")

	default:
		h.logger.Error("failed to update profile", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}

func validateUpdateRequest(updateRequest UpdateRequest) (user.UpdateProfile, error) {
	var update user.UpdateProfile

	if updateRequest.DisplayName != nil {
		displayName := user.NormalizeDisplayName(*updateRequest.DisplayName)
		displayNameLength := utf8.RuneCountInString(displayName)
		if displayNameLength < 2 || displayNameLength > 100 {
			return user.UpdateProfile{}, user.ErrInvalidDisplayName
		}

		update.DisplayName = &displayName
	}

	if updateRequest.Email != nil {
		email := user.NormalizeEmail(*updateRequest.Email)
		if err := user.ValidateEmail(email); err != nil {
			return user.UpdateProfile{}, err
		}

		update.Email = &email
	}

	if update.DisplayName == nil && update.Email == nil {
		return user.UpdateProfile{}, errEmptyUpdate
	}

	return update, nil
}

func writeValidationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, user.ErrInvalidDisplayName):
		writeError(w, http.StatusBadRequest, "INVALID_DISPLAY_NAME", "Le nom doit contenir entre 2 et 100 caracteres.")

	case errors.Is(err, user.ErrInvalidUserEmail):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", "L'adresse email est invalide.")

	case errors.Is(err, errEmptyUpdate):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Au moins un champ doit etre fourni.")

	default:
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "La requete est invalide.")
	}
}

func newResponse(foundUser user.User) Response {
	return Response{
		ID:          foundUser.ID.String(),
		DisplayName: foundUser.DisplayName,
		Email:       foundUser.Email,
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
