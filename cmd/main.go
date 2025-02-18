package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/firebase"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
)

func main() {
	ctx := context.Background()
	cfg, err := config.MustLoad()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	db, err := config.InitDataStore(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	firebaseBucket, err := firebase.InitFirebaseStorage(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect to firebase storage", "error", err)
		return
	}

	services := app.NewServices(db, firebaseBucket)

	router := app.NewRouter(services)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPServer.Port),
		Handler: router,
	}

	serverRunning := make(chan os.Signal, 1)

	signal.Notify(
		serverRunning,
		syscall.SIGABRT,
		syscall.SIGALRM,
		syscall.SIGBUS,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGTSTP,
	)

	go func() {
		slog.Info("server listening at", "port", cfg.HTTPServer.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			serverRunning <- syscall.SIGINT
		}
	}()

	<-serverRunning

	slog.Info("shutting down the server")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("cannot shut HTTP server down gracefully", "error", err)
	}

	slog.Info("server shutdown successfully")
}
