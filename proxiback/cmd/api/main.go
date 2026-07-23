package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/auth"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/database"
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

	authService := auth.NewService(
		userRepository,
		refreshTokenRepository,
		tokenGenerator,
		refreshTokenDurationFromEnv(),
	)

	authHandler := auth.NewHandler(authService, logger)
	profileHandler := profile.NewHandler(userRepository, logger)
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
		"POST /api/v1/auth/forgot-password",
		authHandler.ForgotPassword,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/reset-password",
		authHandler.ResetPassword,
	)

	mux.HandleFunc(
		"PATCH /api/v1/auth/password",
		authHandler.ChangePassword,
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

	registerStaticFrontend(mux, logger)

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

func registerStaticFrontend(mux *http.ServeMux, logger *slog.Logger) {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		return
	}

	indexPath := filepath.Join(staticDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		logger.Warn(
			"frontend static files unavailable",
			"dir",
			staticDir,
			"error",
			err,
		)

		return
	}

	fileServer := http.FileServer(http.Dir(staticDir))
	mux.HandleFunc(
		"GET /",
		func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			cleanPath := strings.TrimPrefix(path.Clean("/"+request.URL.Path), "/")
			filePath := filepath.Join(staticDir, filepath.FromSlash(cleanPath))

			if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
				fileServer.ServeHTTP(responseWriter, request)

				return
			}

			if path.Ext(cleanPath) != "" {
				http.NotFound(responseWriter, request)

				return
			}

			http.ServeFile(responseWriter, request, indexPath)
		},
	)
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
