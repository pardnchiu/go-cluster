package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/pardnchiu/pdcluster/internal/broadcast"
	"github.com/pardnchiu/pdcluster/internal/daemon"
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
	// * check root
	if os.Geteuid() != 0 {
		slog.Error("failed: need to run as root")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("kvmesh [start|status|stop]")
		os.Exit(1)
	}

	action := os.Args[1]

	switch action {
	case "start":
		break
	case "status":
		daemon.Local.Status()
		return
	case "stop":
		if err := daemon.Local.Stop(); err != nil {
			fmt.Printf("failed: stop: %v\n", err)
			os.Exit(1)
		}
		return
	default:
		fmt.Printf("failed: unknown: %q\n", action)
		os.Exit(1)
	}

	// * create daemon
	err := daemon.Local.Init()
	if err != nil {
		slog.Error("failed: init daemon",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer daemon.Local.Stop()

	if err := node.Local.Init(); err != nil {
		slog.Error("failed: init node",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	err = database.Local.Init()
	if err != nil {
		slog.Error("failed: init database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer database.Local.Close()

	ctx, cancel := context.WithCancel(context.Background())
	shutdown(cancel)

	receiver, err := broadcast.InitReceiver(database.Local)
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

func shutdown(cancel context.CancelFunc) {
	chann := make(chan os.Signal, 1)
	signal.Notify(chann, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-chann
		slog.Info("shutting down...")
		cancel()
	}()
}
