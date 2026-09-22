package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/multiplayer-ludo/internal/auth"
	"github.com/multiplayer-ludo/internal/config"
	"github.com/multiplayer-ludo/internal/database"
	"github.com/multiplayer-ludo/internal/middleware"
	"github.com/multiplayer-ludo/internal/room"
	"github.com/multiplayer-ludo/internal/websocket"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database successfully")

	// Repositories
	authRepo := auth.NewRepository(db)
	roomRepo := room.NewRepository(db)

	// Run migrations (create tables if they don't exist)
	if err := authRepo.CreateTable(); err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}
	if err := roomRepo.CreateTables(); err != nil {
		log.Fatalf("failed to create room tables: %v", err)
	}
	log.Println("Database migrations completed")

	// Services
	authService := auth.NewService(authRepo, cfg)

	// Managers
	roomManager := room.NewManager(roomRepo)

	// WebSocket Hub
	hub := websocket.NewHub(roomManager)
	go hub.Run()

	// HTTP Router
	mux := http.NewServeMux()

	// Register routes
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(mux)

	roomHandler := room.NewHandler(roomManager)
	roomHandler.RegisterRoutes(mux, middleware.RequireAuth(cfg))

	wsHandler := websocket.NewHandler(hub, cfg, roomManager)
	wsHandler.RegisterRoutes(mux)

	// Serve demo.html as a static file at root
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/demo.html" {
			http.ServeFile(w, r, "demo.html")
			return
		}
		http.NotFound(w, r)
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🎲 Multiplayer Ludo backend listening on http://localhost%s", addr)
	log.Printf("📄 Open demo client at http://localhost%s/", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
