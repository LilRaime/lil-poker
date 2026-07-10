package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"lil-poker/internal/api"
	"lil-poker/internal/config"
	"lil-poker/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	fmt.Println("Lil-Poker Backend API")
	fmt.Println("=========================================")

	cfg := config.LoadConfig()

	db, err := store.ConnectDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err = store.RunMigrations(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	server := api.NewServer(db, cfg)
	defer server.Close()

	portVal := os.Getenv("PORT")
	if portVal == "" {
		portVal = "8080"
	}
	port := portVal
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	apiHandler := server.Handler()

	distDir := "./dist"
	if _, err := os.Stat(distDir); err != nil {
		if _, errF := os.Stat("./frontend/dist"); errF == nil {
			distDir = "./frontend/dist"
		}
	}
	hasFrontend := false
	if _, err := os.Stat(distDir); err == nil {
		hasFrontend = true
		slog.Info("Serving frontend static files", "dir", distDir)
	}

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			apiHandler.ServeHTTP(w, r)
			return
		}
		if !hasFrontend {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Lil-Poker API Server\nFrontend dist folder not found. Run frontend separately in dev mode (npm run dev)."))
			return
		}
		cleanPath := filepath.Clean(r.URL.Path)
		fullPath := filepath.Join(distDir, cleanPath)
		if _, err := os.Stat(fullPath); err == nil {
			http.FileServer(http.Dir(distDir)).ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		}
	})

	httpServer := &http.Server{
		Addr:    port,
		Handler: mainHandler,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Server starting", "addr", "http://localhost"+port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "err", err)
	}
	slog.Info("Server stopped.")
}
