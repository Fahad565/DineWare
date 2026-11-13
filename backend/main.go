package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dineware-backend/db"
	_ "dineware-backend/docs"
	"dineware-backend/handlers"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           DineWare API Engine
// @version         1.0
// @description     Backend workflow manager handling staggered Africa's Talking dispatches and guest allergy synchronization for safari lodges.
// @termsOfService  http://swagger.io/terms/

// @contact.name    DineWare API Support
// @contact.email   support@dineware.local

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /

// SetupRouter initializes application routes, Swagger UI, and middleware.
func SetupRouter(store db.Store) http.Handler {
	mux := http.NewServeMux()
	h := handlers.NewHandler(store)

	// Health endpoint
	mux.HandleFunc("GET /health", h.Health)

	// Itinerary endpoints
	mux.HandleFunc("POST /api/v1/itineraries", h.CreateItinerary)
	mux.HandleFunc("GET /api/v1/itineraries", h.ListItineraries)
	mux.HandleFunc("GET /api/v1/itineraries/{id}", h.GetItinerary)
	mux.HandleFunc("PATCH /api/v1/itineraries/{id}/guests", h.UpdateGuests)

	// Swagger UI documentation routes
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	return handlers.CORSMiddleware(mux)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var store db.Store
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		log.Println("Connecting to Neon PostgreSQL...")
		dbPool, err := db.Connect(dbURL)
		if err != nil {
			log.Fatalf("Fatal: failed to initialize database: %v", err)
		}
		defer dbPool.Close()
		log.Println("Database connection pool established and migrations applied.")
		store = db.NewPostgresStore(dbPool)
	} else {
		log.Println("Notice: DATABASE_URL not provided. Initializing in-memory fallback store.")
		store = db.NewMemoryStore()
	}

	router := SetupRouter(store)
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("DineWare API Engine running on :%s", port)
		log.Printf("Interactive Swagger UI available at: http://localhost:%s/swagger/index.html", port)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("Initiating shutdown on signal: %v", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Graceful shutdown failed: %v", err)
		}
		log.Println("Server exited cleanly.")
	}
}
