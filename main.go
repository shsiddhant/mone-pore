package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/shsiddhant/mone-pore/internal/db"
	"github.com/shsiddhant/mone-pore/internal/handlers"
)

func main() {

	// Application Configuration
	const (
		appName = "mone-pore"
		port    = "8080"
		timeout = 30 * time.Second
	)

	// Initialize logger
	loggerPrefix := fmt.Sprintf("[%s] ", appName)
	logger := log.New(os.Stdout, loggerPrefix, log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("Starting Mone Pore")

	// Call the app runner
	if err := run(appName, port, timeout, logger); err != nil {
		logger.Fatalf("Application failed to start: %v", err)
	}
}

func run(appName string, port string, timeout time.Duration, logger *log.Logger) error {
	// Create db location
	dir, err := dataDir(appName)
	if err != nil {
		return fmt.Errorf("failed to get data directory: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// DB Configuration
	dbPath := filepath.Join(dir, appName+".db")
	cfg := db.SQLiteConfig{
		DBPath:      dbPath,
		BusyTimeout: 5000,
		JournalMode: "WAL",
		ForeignKeys: true,
		Cache:       "shared",
		Synchronous: "NORMAL",
	}

	// Initialize database
	logger.Println("Initializing SQLite database...")
	database, err := db.Open(cfg)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer func() {
		logger.Println("Closing database cleanly...")
		_ = database.Close()
	}()
	logger.Println("Database initialized successfully")

	// Run migrations
	if err := db.RunMigrations(database, logger); err != nil {
		return fmt.Errorf("migration runner failed: %w", err)
	}

	// Create application instance
	app := &handlers.Application{
		DB: database,
	}

	// Create router using ServerMux
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /{$}", app.Home)

	// Configure server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout,
	}

	// Start server
	logger.Printf("Server starting on http://localhost:%s", port)
	logger.Println("Press Ctrl+C to stop")

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("Server failed to start: %v", err)
	}
	return nil
}

// dataDir returns the directory used to store application data.
func dataDir(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", appName), nil
}
