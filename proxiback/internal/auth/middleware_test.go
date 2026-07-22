package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

func TestAuthenticationMiddlewareInjectsUserID(t *testing.T) {
	t.Parallel()

	generator, err := NewTokenGenerator("secret", "proxibet-api", time.Hour)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error = %v", err)
	}

	authenticatedUser := user.User{
		ID:          uuid.New(),
		DisplayName: "Martin",
		Email:       "martin@example.com",
	}

	token, _, err := generator.Generate(authenticatedUser)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("user id not found in context")
		}

		if userID != authenticatedUser.ID {
			t.Fatalf("userID = %s, want %s", userID, authenticatedUser.ID)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	NewAuthenticationMiddleware(generator)(next).ServeHTTP(response, request)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestAuthenticationMiddlewareRejectsMissingBearerToken(t *testing.T) {
	t.Parallel()

	generator, err := NewTokenGenerator("secret", "proxibet-api", time.Hour)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error = %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	response := httptest.NewRecorder()

	NewAuthenticationMiddleware(generator)(next).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
