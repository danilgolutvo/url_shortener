package main

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"url_shortener/cmd/middleware"
	"url_shortener/httpServer/handlers/deleteURL"
	"url_shortener/httpServer/handlers/login"
	"url_shortener/httpServer/handlers/redirect"
	"url_shortener/httpServer/handlers/register"
	"url_shortener/httpServer/handlers/url/save"
	"url_shortener/internal/config"
	"url_shortener/internal/lib/logger/handlers/slogpretty"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO: init config - library - cleanenv
	// Create dir "config" in the root with local.yaml and store parameters of config, create dir "internal" and within it dir "config" with config.go file and create structs fitting for storage of local.yaml parameters, use library cleanenv to read config and put it in the created structs, use export CONFIG_PATH=/Users/dangolutvo/Documents/GitHub/url_shortener/config/local.yaml

	ctx := context.Background()
	cfg := config.MustLoad()

	// TODO: init logger - library - sl (import log/sl)
	// make setupLogger func that defines which env i'm in to create a logger with corresponding Level, create log.Info with two args, the latter is for understanding which env I'm currently in

	log := setupLogger(cfg.Env)
	log.Info(
		"starting url_shortener",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	// TODO: init storage - library - postgres or SQLite
	// create dir "storage" within dir "internal" with storage.go and dir "postgres" with postgres.go
	storage, err := postgres.NewStorage(ctx, cfg.Dsn)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer storage.DB.Close()
	log.Info("Database initialized successfully")

	// TODO: init router - library - chi, chi"render" or gorilla
	router := mux.NewRouter()

	// middleware that attaches uniq id to a request
	router.Use(middleware.RequestID)
	router.Use(middleware.LoggingMiddleware)
	router.Handle("/{alias}", redirect.New(log, storage)).Methods(http.MethodGet)
	router.Handle("/login", login.HandleLogin(log, storage)).Methods(http.MethodPost)
	router.Handle("/register", register.HandleRegistration(log, storage)).Methods(http.MethodPost)

	privateRouter := router.PathPrefix("/").Subrouter()
	privateRouter.Use(middleware.Auth)

	privateRouter.Handle("/url", save.New(log, storage)).Methods(http.MethodPost)
	privateRouter.Handle("/url/{alias}", deleteURL.New(log, storage)).Methods(http.MethodDelete)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the request ID from the context
		reqID := middleware.GetReqID(r.Context())
		fmt.Fprintf(w, "Hello, your request ID is: %s\n", reqID)
	})
	recoveryHandler := handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true), // Print stack trace to logs
		handlers.RecoveryLogger(slog.NewLogLogger(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}), // Log to stdout
			slog.LevelError, // Log level for errors
		)),
	)(router)

	log.Info("Starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      recoveryHandler,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("SHOULD NOT SEE THIS MESSAGE")

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
