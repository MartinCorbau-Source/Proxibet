package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

type handler struct {
	authService *Service
	logger      *slog.Logger
}

func NewHandler(authService *Service, logger *slog.Logger) *handler {
	return &handler{
		authService: authService,
		logger:      logger,
	}
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&registerRequest); err != nil {
		h.logger.Error("failed to decode request", err)
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request")
		return
	}
    response, err := h.authService.Register(r.Context(), registerRequest.Email, registerRequest.Password, registerRequest.DisplayName)
	switch {
	case err == nil:
		writeJSON(
			w,
			http.StatusCreated,
			response,
		)

	case errors.Is(err, user.ErrEmailAlreadyInUse):
		writeError(
			w,
			http.StatusConflict,
			"EMAIL_ALREADY_USED",
			"Cette adresse email est déjà utilisée.",
		)

	case errors.Is(err, ErrInvalidEmailFormat):
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_EMAIL",
			"L'adresse email est invalide.",
		)

	case errors.Is(err, ErrInvalidDisplayName):
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_DISPLAY_NAME",
			"Le nom doit contenir entre 2 et 100 caractères.",
		)

	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooLong):
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_PASSWORD",
			"Le mot de passe doit contenir entre 8 et 72 caractères.",
		)

	default:
		h.logger.Error(
			"failed to register user",
			err,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Une erreur interne est survenue.",
		)
	}
}


func writeJSON(
	responseWriter http.ResponseWriter,
	status int,
	value any,
) {
	responseWriter.Header().Set(
		"Content-Type",
		"application/json",
	)
	responseWriter.WriteHeader(status)

	_ = json.NewEncoder(responseWriter).Encode(value)
}

func writeError(w http.ResponseWriter,status int,code string, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}