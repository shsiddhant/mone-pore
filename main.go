package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"

	"github.com/shsiddhant/mone-pore/internal/db"
	"github.com/shsiddhant/mone-pore/internal/handlers"
	"github.com/shsiddhant/mone-pore/internal/middleware"
)

//go:embed ui/static
var staticFS embed.FS

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

	// Initialize a new session manager
	sessionManager := scs.New()

	// Configure session store, lifetime, and idle timeout.
	sessionManager.Store = sqlite3store.New(database.DB)
	sessionManager.Lifetime = 30 * time.Minute
	sessionManager.IdleTimeout = 15 * time.Minute

	// Configure cookies
	sessionManager.Cookie.Name = appName + "-session"

	// Create application instance
	app := &handlers.Application{
		DB:             database,
		SessionManager: sessionManager,
	}

	// Create router using ServerMux
	mux := http.NewServeMux()

	// Serve static files from embedded filesystem
	staticSubFS, err := fs.Sub(staticFS, "ui/static")
	if err != nil {
		logger.Fatalf("Failed to create static sub-filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticSubFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Routes
	mux.HandleFunc("GET /{$}", app.Home)
	mux.HandleFunc("GET /journal/{id}/unlock", app.UnlockJournalForm)
	mux.HandleFunc("POST /journal/{id}/unlock", app.UnlockJournal)
	mux.HandleFunc("POST /journal/{id}/lock", app.LockJournal)
	mux.HandleFunc("GET /journal/new", app.NewJournalPage)
	mux.HandleFunc("POST /journal/new", app.NewJournal)

	unlockMiddleware := middleware.UnlockedJournalRequired(app.SessionManager)

	// Unlocked Journal Required.
	mux.Handle(
		"GET /journal/{id}",
		unlockMiddleware.Apply(app.JournalIndex),
	)
	mux.Handle(
		"GET /journal/{id}/memory/new", // NewMemoryPage: GET
		unlockMiddleware.Apply(app.NewMemoryPage),
	)
	mux.Handle(
		"POST /journal/{id}/memory/new", // NewMemory: POST
		unlockMiddleware.Apply(app.NewMemory),
	)
	mux.Handle(
		"GET /journal/{id}/memory/{memory_id}",
		unlockMiddleware.Apply(app.MemoryIndex), // MemoryIndex: GET
	)
	mux.Handle(
		"GET /journal/{id}/export_json",
		unlockMiddleware.Apply(app.ExportJournalToJSON),
	)
	mux.Handle(
		"GET /journal/{id}/memory/{memory_id}/edit",
		unlockMiddleware.Apply(app.EditMemoryPage),
	)
	mux.Handle(
		"POST /journal/{id}/memory/{memory_id}/edit",
		unlockMiddleware.Apply(app.EditMemory),
	)

	mux.HandleFunc("GET /admin/setup", app.SetupAdminPage)
	mux.HandleFunc("POST /admin/setup", app.SetupAdmin)

	// Global middlewares
	withCSRF := middleware.CSRF(mux)
	withSession := app.SessionManager.LoadAndSave(withCSRF)
	withAdminSetup := middleware.AdminSetupRequired(
		app.DB.AdminPasswordSetup)(withSession)

	// Configure server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      withAdminSetup,
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
