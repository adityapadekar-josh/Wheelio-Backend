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
)

func main() {
	ctx := context.Background()

	cfg, err := config.MustLoad()
	if err != nil {
		slog.Error("Failed to load config", "error", err.Error())
		return
	}

	db, err := config.InitDataStore(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err.Error())
		return
	}
	defer db.Close()

	services := app.NewServices(db)

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
		slog.Info("Server listening at", "port", cfg.HTTPServer.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", slog.String("error", err.Error()))
			serverRunning <- syscall.SIGINT
		}
	}()

	<-serverRunning

	slog.Info("Shutting down the server")
	ctxWT, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxWT); err != nil {
		slog.Error("Cannot shut HTTP server down gracefully", "error", err.Error())
	}

	slog.Info("Server shutdown successfully")
}
