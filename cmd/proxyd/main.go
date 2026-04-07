package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	apiv1 "github.com/ryfineZ/weave/internal/api/core/v1"
	"github.com/ryfineZ/weave/internal/log"
)

// Injected at build time via -ldflags.
var (
	version   = "0.1.0"
	commit    = "dev"
	buildDate = "unknown"
)

const (
	defaultSocketPath = "/var/run/weave/proxyd.sock"
	defaultLogDir     = "/Library/Logs/Weave"
	shutdownTimeout   = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "weave daemon: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		socketPath = flag.String("socket", defaultSocketPath, "Unix domain socket path")
		logDir     = flag.String("log-dir", defaultLogDir, "Directory for log files")
		debug      = flag.Bool("debug", false, "Enable debug logging and stderr output")
		showVer    = flag.Bool("version", false, "Print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("weave daemon %s (%s) built %s\n", version, commit, buildDate)
		return nil
	}

	logger, err := log.New(*logDir, *debug)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("weave daemon starting",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("socket", *socketPath),
	)

	// Ensure socket directory exists and remove stale socket.
	if err := os.MkdirAll(filepath.Dir(*socketPath), 0o750); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}
	_ = os.Remove(*socketPath)

	listener, err := net.Listen("unix", *socketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", *socketPath, err)
	}
	// Restrict socket to root + group (group will be set to staff/admin for UI access).
	if err := os.Chmod(*socketPath, 0o660); err != nil {
		return fmt.Errorf("chmod socket: %w", err)
	}

	srv := &http.Server{
		Handler:           apiv1.NewServer(logger),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown on SIGTERM / SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("daemon listening", zap.String("socket", *socketPath))
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining connections…")
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}

	// Clean up socket file.
	_ = os.Remove(*socketPath)

	logger.Info("daemon stopped cleanly")
	return nil
}
