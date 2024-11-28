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
	"url_shortener/internal/config"
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
	log.Info("starting url_shortener", slog.String("env", cfg.Env))
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

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the request ID from the context
		reqID := middleware.GetReqID(r.Context())
		fmt.Fprintf(w, "Hello, your request ID is: %s\n", reqID)
	})

	slog.Info("Starting server", slog.String("address", ":8084"))
	http.ListenAndServe(":8084", handlers.RecoveryHandler()(router))

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	if env == envLocal {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	if env == envDev {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	if env == envProd {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
