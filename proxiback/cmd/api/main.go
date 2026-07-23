package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/auth"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/database"
	groups "github.com/MartinCorbau-Source/proxibet/proxiback/internal/group"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/httpx"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/profile"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
	"github.com/MartinCorbau-Source/proxibet/proxiback/migrations"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is missing")
		os.Exit(1)
	}

	pool, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)
	if err != nil {
		logger.Error(
			"could not create database pool",
			"error",
			err,
		)

		os.Exit(1)
	}
	defer pool.Close()

	tokenGenerator, err := auth.NewTokenGenerator(
		os.Getenv("JWT_SECRET"),
		os.Getenv("JWT_ISSUER"),
		accessTokenDurationFromEnv(),
	)
	if err != nil {
		logger.Error(
			"could not create token generator",
			"error",
			err,
		)

		os.Exit(1)
	}

	if err := pool.Ping(context.Background()); err != nil {
		logger.Error(
			"could not connect to database",
			"error",
			err,
		)

		os.Exit(1)
	}

	if err := database.Migrate(context.Background(), pool, migrations.Files, logger); err != nil {
		logger.Error(
			"could not apply database migrations",
			"error",
			err,
		)

		os.Exit(1)
	}

	userRepository := user.NewPostgresRepository(pool)
	refreshTokenRepository := auth.NewPostgresRefreshTokenRepository(pool)
	groupRepository := groups.NewPostgresRepository(pool)

	authService := auth.NewService(
		userRepository,
		refreshTokenRepository,
		tokenGenerator,
		refreshTokenDurationFromEnv(),
	)

	groupService := groups.NewService(groupRepository)

	authHandler := auth.NewHandler(authService, logger)
	profileHandler := profile.NewHandler(userRepository, logger)
	groupHandler := groups.NewHandler(groupService, logger)
	requireAuth := auth.NewAuthenticationMiddleware(tokenGenerator)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
		func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			responseWriter.WriteHeader(http.StatusOK)
			_, _ = responseWriter.Write([]byte(`{"status":"ok"}`))
		},
	)

	mux.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/refresh",
		authHandler.Refresh,
	)

	mux.Handle(
		"GET /api/v1/me",
		requireAuth(http.HandlerFunc(profileHandler.GetMe)),
	)

	mux.Handle(
		"PATCH /api/v1/me",
		requireAuth(http.HandlerFunc(profileHandler.UpdateMe)),
	)

	mux.Handle(
		"POST /api/v1/groups",
		requireAuth(http.HandlerFunc(groupHandler.CreateGroup)),
	)

	mux.Handle(
		"GET /api/v1/groups",
		requireAuth(http.HandlerFunc(groupHandler.ListGroups)),
	)

	mux.Handle(
		"GET /api/v1/groups/{groupId}",
		requireAuth(http.HandlerFunc(groupHandler.GetGroup)),
	)

	mux.Handle(
		"DELETE /api/v1/groups/{groupId}",
		requireAuth(http.HandlerFunc(groupHandler.DeleteGroup)),
	)

	mux.Handle(
		"POST /api/v1/groups/{groupId}/invitations",
		requireAuth(http.HandlerFunc(groupHandler.CreateInvitation)),
	)

	mux.Handle(
		"PATCH /api/v1/groups/{groupId}/invitations/{invitationId}",
		requireAuth(http.HandlerFunc(groupHandler.DeactivateInvitation)),
	)

	mux.Handle(
		"POST /api/v1/groups/join",
		requireAuth(http.HandlerFunc(groupHandler.JoinGroup)),
	)

	mux.Handle(
		"GET /api/v1/groups/{groupId}/members",
		requireAuth(http.HandlerFunc(groupHandler.ListMembers)),
	)

	mux.Handle(
		"PATCH /api/v1/groups/{groupId}/members/{userId}",
		requireAuth(http.HandlerFunc(groupHandler.UpdateMember)),
	)

	mux.Handle(
		"DELETE /api/v1/groups/{groupId}/members/{userId}",
		requireAuth(http.HandlerFunc(groupHandler.DeleteMember)),
	)
	server := &http.Server{
		Addr:    ":8080",
		Handler: httpx.CORSMiddleware(corsAllowedOriginsFromEnv())(mux),
	}

	logger.Info("starting API", "port", 8080)

	if err := server.ListenAndServe(); err != nil {
		logger.Error(
			"HTTP server stopped",
			"error",
			err,
		)

		os.Exit(1)
	}
}

func corsAllowedOriginsFromEnv() []string {
	corsAllowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsAllowedOrigins == "" {
		return []string{"http://localhost:6767"}
	}

	origins := strings.Split(corsAllowedOrigins, ",")
	allowedOrigins := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins = append(allowedOrigins, origin)
		}
	}

	if len(allowedOrigins) == 0 {
		return []string{"http://localhost:6767"}
	}

	return allowedOrigins
}

func accessTokenDurationFromEnv() time.Duration {
	if duration := durationFromEnv("ACCESS_TOKEN_DURATION"); duration > 0 {
		return duration
	}

	return durationFromEnv("JWT_DURATION")
}

func refreshTokenDurationFromEnv() time.Duration {
	if duration := durationFromEnv("REFRESH_TOKEN_DURATION"); duration > 0 {
		return duration
	}

	return 30 * 24 * time.Hour
}

func durationFromEnv(name string) time.Duration {
	durationStr := os.Getenv(name)
	if durationStr == "" {
		return 0
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0
	}

	return duration
}
