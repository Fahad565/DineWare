package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dineware-backend/db"
	_ "dineware-backend/docs"
	"dineware-backend/handlers"
	"dineware-backend/integrations"
	"dineware-backend/worker"

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

	if ussdStore, ok := store.(ussdDataStore); ok {
		mux.Handle("POST /api/v1/ussd", integrations.USSDWebhook(ussdProvider{store: ussdStore}))
	} else {
		mux.HandleFunc("POST /api/v1/ussd", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			http.Error(w, "END USSD service unavailable", http.StatusServiceUnavailable)
		})
	}

	// Swagger UI documentation routes
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	return handlers.CORSMiddleware(mux)
}

type ussdDataStore interface {
	GetStopByItineraryAndPhone(context.Context, string, string) (*db.USSDStop, error)
	GetGuestsByItinerary(context.Context, string) ([]db.USSDGuest, error)
	ConfirmManifest(context.Context, string, string) error
}

type ussdProvider struct {
	store ussdDataStore
}

func (p ussdProvider) GetStopByItineraryAndPhone(ctx context.Context, itineraryID, chefPhone string) (*integrations.StopRecord, error) {
	stop, err := p.store.GetStopByItineraryAndPhone(ctx, itineraryID, chefPhone)
	if err != nil || stop == nil {
		return nil, err
	}
	return &integrations.StopRecord{
		ItineraryID: stop.ItineraryID, GroupName: stop.GroupName, LodgeName: stop.LodgeName,
		ChefPhone: stop.ChefPhone, CheckInDate: stop.CheckInDate,
		GuestCount: stop.GuestCount, ChefConfirmed: stop.ChefConfirmed,
	}, nil
}

func (p ussdProvider) GetGuestsByItinerary(ctx context.Context, itineraryID string) ([]integrations.GuestRecord, error) {
	guests, err := p.store.GetGuestsByItinerary(ctx, itineraryID)
	if err != nil {
		return nil, err
	}
	result := make([]integrations.GuestRecord, len(guests))
	for i, guest := range guests {
		result[i] = integrations.GuestRecord{
			Name: guest.Name, Allergies: guest.Allergies, DietaryRequirements: guest.DietaryRequirements,
		}
	}
	return result, nil
}

func (p ussdProvider) ConfirmManifest(ctx context.Context, itineraryID, chefPhone string) error {
	return p.store.ConfirmManifest(ctx, itineraryID, chefPhone)
}

type dispatchDataStore interface {
	QueryPendingStops(context.Context, time.Duration) ([]db.PendingStop, error)
	MarkNotificationSent(context.Context, int) error
}

type stopRepository struct {
	store dispatchDataStore
}

func (r stopRepository) QueryPendingStops(ctx context.Context, within time.Duration) ([]worker.PendingStop, error) {
	stops, err := r.store.QueryPendingStops(ctx, within)
	if err != nil {
		return nil, err
	}
	result := make([]worker.PendingStop, len(stops))
	for i, stop := range stops {
		result[i] = worker.PendingStop{
			StopID: stop.StopID, ItineraryID: stop.ItineraryID, GroupName: stop.GroupName,
			ChefPhone: stop.ChefPhone, LodgeName: stop.LodgeName,
			CheckInDate: stop.CheckInDate, GuestCount: stop.GuestCount,
		}
	}
	return result, nil
}

func (r stopRepository) MarkNotificationSent(ctx context.Context, stopID int) error {
	return r.store.MarkNotificationSent(ctx, stopID)
}

func runDispatcher(ctx context.Context, dispatcher *worker.Dispatcher) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		if err := dispatcher.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("SMS dispatcher cycle failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
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

	atClient := integrations.NewATClient(os.Getenv("AT_USERNAME"), os.Getenv("AT_API_KEY"))
	workerCtx, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan struct{})
	if repository, ok := store.(dispatchDataStore); ok {
		dispatcher := worker.NewDispatcher(stopRepository{store: repository}, worker.NewATSMSSender(atClient), log.Default())
		go func() {
			defer close(workerDone)
			runDispatcher(workerCtx, dispatcher)
		}()
	} else {
		log.Println("SMS dispatcher is unavailable: configured store does not support stop queries")
		close(workerDone)
	}
	defer stopWorker()

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
	defer signal.Stop(shutdown)

	select {
	case err := <-serverErrors:
		stopWorker()
		<-workerDone
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("Initiating shutdown on signal: %v", sig)
		stopWorker()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			stopWorker()
			<-workerDone
			log.Fatalf("Graceful shutdown failed: %v", err)
		}
		<-workerDone
		log.Println("Server exited cleanly.")
	}
}
