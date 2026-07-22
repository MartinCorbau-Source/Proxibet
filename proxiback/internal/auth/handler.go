package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

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
		h.logger.Error("failed to decode request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request")
		return
	}

	response, err := h.authService.Register(r.Context(), registerRequest.Email, registerRequest.Password, registerRequest.DisplayName)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, response)

	case errors.Is(err, user.ErrEmailAlreadyInUse), errors.Is(err, user.ErrUserAlreadyExists):
		writeError(w, http.StatusConflict, "EMAIL_ALREADY_USED", "Cette adresse email est deja utilisee.")

	case errors.Is(err, ErrInvalidEmailFormat):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", "L'adresse email est invalide.")

	case errors.Is(err, ErrInvalidDisplayName):
		writeError(w, http.StatusBadRequest, "INVALID_DISPLAY_NAME", "Le nom doit contenir entre 2 et 100 caracteres.")

	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooLong), errors.Is(err, ErrInvalidPasswordFormat):
		writeError(w, http.StatusBadRequest, "INVALID_PASSWORD", "Le mot de passe doit contenir entre 8 et 72 caracteres.")

	default:
		h.logger.Error("failed to register user", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&loginRequest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Le corps de la requete est invalide.")
		return
	}

	response, err := h.authService.Login(r.Context(), loginRequest.Email, loginRequest.Password)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, response)

	case errors.Is(err, user.ErrUserNotFound), errors.Is(err, user.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "L'adresse email ou le mot de passe est incorrect.")

	case errors.Is(err, ErrInvalidPasswordFormat):
		writeError(w, http.StatusBadRequest, "INVALID_PASSWORD", "Le mot de passe doit contenir entre 8 et 72 caracteres.")

	case errors.Is(err, user.ErrAccountInactive):
		writeError(w, http.StatusUnauthorized, "ACCOUNT_DISABLED", "Votre compte a ete desactive.")

	default:
		h.logger.Error("failed to login user", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}

func (h *handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshRequest RefreshRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&refreshRequest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Le corps de la requete est invalide.")
		return
	}

	response, err := h.authService.Refresh(r.Context(), refreshRequest.RefreshToken)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, response)

	case errors.Is(err, ErrInvalidRefreshToken), errors.Is(err, ErrRefreshTokenExpired), errors.Is(err, ErrRefreshTokenRevoked), errors.Is(err, user.ErrUserNotFound):
		writeError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Le refresh token est invalide ou expire.")

	case errors.Is(err, user.ErrAccountInactive):
		writeError(w, http.StatusUnauthorized, "ACCOUNT_DISABLED", "Votre compte a ete desactive.")

	default:
		h.logger.Error("failed to refresh token", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}

func (h *handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var changePasswordRequest ChangePasswordRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&changePasswordRequest); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Le corps de la requete est invalide.")
		return
	}

	accessToken, err := bearerTokenFromHeader(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Vous devez etre connecte pour changer votre mot de passe.")
		return
	}

	err = h.authService.ChangePassword(
		r.Context(),
		accessToken,
		changePasswordRequest.CurrentPassword,
		changePasswordRequest.NewPassword,
	)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)

	case errors.Is(err, user.ErrInvalidToken):
		writeError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Vous devez etre connecte pour changer votre mot de passe.",
		)

	case errors.Is(err, user.ErrTokenExpired):
		writeError(
			w,
			http.StatusUnauthorized,
			"TOKEN_EXPIRED",
			"Votre session a expire.",
		)

	case errors.Is(err, user.ErrUserNotFound):
		writeError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Vous devez etre connecte pour changer votre mot de passe.",
		)

	case errors.Is(err, user.ErrInvalidCredentials):
		writeError(
			w,
			http.StatusUnauthorized,
			"INVALID_CREDENTIALS",
			"Le mot de passe actuel est incorrect.",
		)

	case errors.Is(err, user.ErrAccountInactive):
		writeError(
			w,
			http.StatusUnauthorized,
			"ACCOUNT_DISABLED",
			"Votre compte a ete desactive.",
		)

	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooLong):
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_PASSWORD",
			"Le mot de passe doit contenir entre 8 et 72 caracteres.",
		)

	default:
		h.logger.Error("failed to change password", "error", err)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Une erreur interne est survenue.",
		)
	}
}

func (h *handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var forgotPasswordRequest ForgotPasswordRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&forgotPasswordRequest); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Le corps de la requete est invalide.")
		return
	}

	_, err := h.authService.RequestPasswordReset(r.Context(), forgotPasswordRequest.Email)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)

	case errors.Is(err, ErrInvalidEmailFormat):
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_EMAIL",
			"L'adresse email est invalide.",
		)

	default:
		h.logger.Error("failed to request password reset", "error", err)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Une erreur interne est survenue.",
		)
	}
}

func (h *handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var resetPasswordRequest ResetPasswordRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&resetPasswordRequest); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Le corps de la requete est invalide.")
		return
	}

	err := h.authService.ResetPassword(
		r.Context(),
		resetPasswordRequest.Token,
		resetPasswordRequest.NewPassword,
	)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)

	case errors.Is(err, user.ErrInvalidPasswordResetToken):
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_RESET_TOKEN",
			"Le lien de reinitialisation est invalide ou expire.",
		)

	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooLong):
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_PASSWORD",
			"Le mot de passe doit contenir entre 8 et 72 caracteres.",
		)

	default:
		h.logger.Error("failed to reset password", "error", err)

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

func bearerTokenFromHeader(authorizationHeader string) (string, error) {
	authorizationHeader = strings.TrimSpace(authorizationHeader)

	scheme, token, found := strings.Cut(authorizationHeader, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", user.ErrInvalidToken
	}

	token = strings.TrimSpace(token)
	if token == "" || strings.Contains(token, " ") {
		return "", user.ErrInvalidToken
	}

	return token, nil
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}
