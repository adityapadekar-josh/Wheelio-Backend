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

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/api"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		slog.Error("failed to load config", "error", err.Error())
		return
	}

	db, err := repository.InitDataStore(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err.Error())
		return
	}
	defer db.Close()

	services := app.NewServices(db, cfg)

	router := api.NewRouter(services)

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
			slog.Error("server error", "error", err.Error())
			serverRunning <- syscall.SIGINT
		}
	}()

	<-serverRunning

	slog.Info("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("cannot shut HTTP server down gracefully", "error", err.Error())
	}

	slog.Info("server shutdown successfully")
}
