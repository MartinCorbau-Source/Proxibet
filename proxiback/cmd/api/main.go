package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/auth"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/httpx"
	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
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
		jwtDurationFromEnv(),
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

	userRepository := user.NewPostgresRepository(pool)

	authService := auth.NewService(
		userRepository,
		tokenGenerator,
	)

	authHandler := auth.NewHandler(authService, logger)

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

	allowedOrigins := []string{"http://localhost:6767"}
	if corsAllowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsAllowedOrigins != "" {
		allowedOrigins = strings.Split(corsAllowedOrigins, ",")
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: httpx.CORSMiddleware(allowedOrigins)(mux),
	}

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

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

func jwtDurationFromEnv() time.Duration {
	durationStr := os.Getenv("JWT_DURATION")
	if durationStr == "" {
		return 0
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0
	}

	return duration
}
