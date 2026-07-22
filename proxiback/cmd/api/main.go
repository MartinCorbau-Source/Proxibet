package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := getEnvironmentVariable("HTTP_PORT", "8080")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusOK)

		_, _ = responseWriter.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"starting HTTP server",
			"port",
			port,
		)

		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal := make(chan os.Signal, 1)

	signal.Notify(
		shutdownSignal,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case signal := <-shutdownSignal:
		logger.Info(
			"shutdown signal received",
			"signal",
			signal.String(),
		)

	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"HTTP server stopped unexpectedly",
				"error",
				err,
			)

			os.Exit(1)
		}
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error(
			"could not gracefully stop HTTP server",
			"error",
			err,
		)

		os.Exit(1)
	}

	logger.Info("HTTP server stopped")
}

func getEnvironmentVariable(name string, defaultValue string) string {
	value := os.Getenv(name)

	if value == "" {
		return defaultValue
	}

	return value
}

func buildAddress(host string, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}