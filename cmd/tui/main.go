package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/pardnchiu/pdcluster/internal/broadcast"
	"github.com/pardnchiu/pdcluster/internal/database"
	"github.com/pardnchiu/pdcluster/internal/node"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Error loading .env file",
			slog.String("error", err.Error()))
	}
}

func main() {
	if os.Geteuid() != 0 {
		slog.Error("failed: need to run as root")
		os.Exit(1)
	}

	node.InitKeyPair()
	node.InitCluster()

	db, err := database.InitSQLite()
	if err != nil {
		slog.Error("failed: init database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	shurdown(cancel)

	receiver, err := broadcast.InitReceiver(db)
	if err != nil {
		slog.Error("failed: init receiver",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer receiver.Close()

	go func() {
		if err := receiver.Start(ctx); err != nil && err != context.Canceled {
			slog.Error("receiver error", slog.String("error", err.Error()))
		}
	}()

	sender, err := broadcast.InitSender()
	if err != nil {
		slog.Error("failed: init sender",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer sender.Close()

	go func() {
		if err := sender.Start(ctx); err != nil && err != context.Canceled {
			slog.Error("broadcaster error", slog.String("error", err.Error()))
		}
	}()

	slog.Info("cluster daemon started", slog.Int("peers", 1))
	<-ctx.Done()

	time.Sleep(100 * time.Millisecond)
	slog.Info("cluster daemon stopped")
}

func shurdown(cancel context.CancelFunc) {
	chann := make(chan os.Signal, 1)
	signal.Notify(chann, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-chann
		slog.Info("shutting down...")
		cancel()
	}()
}
