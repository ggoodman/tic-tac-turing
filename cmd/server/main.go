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

	"github.com/ggoodman/tic-tac-turing/internal/mcp"
	"github.com/ggoodman/tic-tac-turing/internal/web"
	"github.com/joeshaw/envdecode"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var cfg Config

	if err := envdecode.Decode(&cfg); err != nil {
		log.ErrorContext(ctx, "failed to decode config from environment", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Initialize web content (parses markdown at startup)
	web.Init()

	mcpUrl := cfg.PublicUrl + "/mcp"

	mcpHandler, err := mcp.NewTicTacTuringHandler(ctx, log, mcpUrl, cfg.AuthIssuerUrl, cfg.RedisUrl)
	if err != nil {
		log.ErrorContext(ctx, "failed to create MCP handler", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Create serve mux
	mux := http.NewServeMux()

	// Register static CSS file (higher priority)
	mux.HandleFunc("/styles.css", web.StylesHandler)

	// Explicit favicon & manifest assets (served from embedded FS)
	mux.HandleFunc("/favicon.ico", web.StaticAssetHandler)
	mux.HandleFunc("/favicon-16x16.png", web.StaticAssetHandler)
	mux.HandleFunc("/favicon-32x32.png", web.StaticAssetHandler)
	mux.HandleFunc("/apple-touch-icon.png", web.StaticAssetHandler)
	mux.HandleFunc("/android-chrome-192x192.png", web.StaticAssetHandler)
	mux.HandleFunc("/android-chrome-512x512.png", web.StaticAssetHandler)
	mux.HandleFunc("/site.webmanifest", web.StaticAssetHandler)

	// Register web handler for root and markdown variants
	mux.HandleFunc("GET /{$}", web.Handler)
	mux.HandleFunc("GET /home.md", web.Handler)
	mux.HandleFunc("GET /index.md", web.Handler)

	// Register MCP handler as fallback - handles /mcp and .well-known paths
	mux.Handle("/", mcpHandler)

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
		// ReadTimeout:  15 * time.Second,
		// WriteTimeout: 15 * time.Second,
		// IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.InfoContext(ctx, "server started", slog.Int("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.ErrorContext(ctx, "error starting server", slog.Int("port", cfg.Port), slog.String("err", err.Error()))
		}
	}()

	<-ctx.Done()

	log.InfoContext(ctx, "shutting down server", slog.String("reason", ctx.Err().Error()))

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.ErrorContext(ctx, "server forced to shut down", slog.String("err", err.Error()))
		os.Exit(1)
	}

	log.InfoContext(ctx, "server exited properly")
}
