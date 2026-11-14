package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"dineware-backend/models"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

var (
	ErrNotFound       = errors.New("record not found")
	ErrEmptyGroupName = errors.New("group_name is required")
	ErrEmptyStops     = errors.New("at least one stop is required")
	ErrEmptyGuests    = errors.New("at least one guest is required")
)

// Store provides an abstraction for data persistence operations.
type Store interface {
	CreateItinerary(ctx context.Context, req models.CreateItineraryRequest) (*models.Itinerary, error)
	ListItineraries(ctx context.Context) ([]models.Itinerary, error)
	GetItinerary(ctx context.Context, id string) (*models.Itinerary, error)
	UpdateGuests(ctx context.Context, itineraryID string, guests []models.GuestInput) (*models.Itinerary, error)
	Ping(ctx context.Context) error
	Close() error
}

// USSDStop and USSDGuest contain the database projections needed by the USSD handler.
type USSDStop struct {
	ItineraryID   string
	GroupName     string
	LodgeName     string
	ChefPhone     string
	CheckInDate   string
	GuestCount    int
	ChefConfirmed bool
}

type USSDGuest struct {
	Name                string
	Allergies           string
	DietaryRequirements string
}

// PendingStop contains only the fields needed to send a privacy-safe notification.
type PendingStop struct {
	StopID      int
	ItineraryID string
	GroupName   string
	ChefPhone   string
	LodgeName   string
	CheckInDate string
	GuestCount  int
}

// Connect establishes a pooled connection to Neon PostgreSQL and executes migrations.
func Connect(databaseURL string) (*sql.DB, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("DATABASE_URL is empty")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection pooling configuration optimized for Neon PostgreSQL
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Automatic migration execution on startup
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// RunMigrations applies all embedded SQL migration scripts in sorted order.
func RunMigrations(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure schema_migrations table exists
	createMigrationTable := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.ExecContext(ctx, createMigrationTable); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		filename := entry.Name()

		// Check if migration was already executed
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`
		if err := db.QueryRowContext(ctx, query, filename).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check migration %s: %w", filename, err)
		}

		if exists {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin migration transaction for %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		recordSQL := `INSERT INTO schema_migrations (version) VALUES ($1)`
		if _, err := tx.ExecContext(ctx, recordSQL, filename); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("Applied migration: %s", filename)
	}

	return nil
}

// PostgresStore implements Store backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgresStore instance.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) CreateItinerary(ctx context.Context, req models.CreateItineraryRequest) (*models.Itinerary, error) {
	if strings.TrimSpace(req.GroupName) == "" {
		return nil, ErrEmptyGroupName
	}
	if len(req.Stops) == 0 {
		return nil, ErrEmptyStops
	}
	if len(req.Guests) == 0 {
		return nil, ErrEmptyGuests
	}

	itinID := strings.TrimSpace(req.ID)
	if itinID == "" {
		itinID = fmt.Sprintf("ITIN-%d", time.Now().UnixNano()%10000000)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var createdAt, updatedAt time.Time
	queryItin := `
		INSERT INTO itineraries (id, group_name, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING created_at, updated_at`
	if err := tx.QueryRowContext(ctx, queryItin, itinID, req.GroupName).Scan(&createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("insert itinerary: %w", err)
	}

	stops := make([]models.Stop, 0, len(req.Stops))
	for _, st := range req.Stops {
		var stopID int
		var notifSent, chefConf bool
		queryStop := `
			INSERT INTO itinerary_stops (itinerary_id, lodge_name, chef_phone, check_in_date)
			VALUES ($1, $2, $3, $4)
			RETURNING id, notification_sent, chef_confirmed`
		if err := tx.QueryRowContext(ctx, queryStop, itinID, st.LodgeName, st.ChefPhone, st.CheckInDate).Scan(&stopID, &notifSent, &chefConf); err != nil {
			return nil, fmt.Errorf("insert stop: %w", err)
		}
		stops = append(stops, models.Stop{
			ID:               stopID,
			ItineraryID:      itinID,
			LodgeName:        st.LodgeName,
			ChefPhone:        st.ChefPhone,
			CheckInDate:      st.CheckInDate,
			NotificationSent: notifSent,
			ChefConfirmed:    chefConf,
		})
	}

	guests := make([]models.Guest, 0, len(req.Guests))
	for _, g := range req.Guests {
		var guestID int
		queryGuest := `
			INSERT INTO guests (itinerary_id, name, allergies, dietary_requirements)
			VALUES ($1, $2, $3, $4)
			RETURNING id`
		if err := tx.QueryRowContext(ctx, queryGuest, itinID, g.Name, g.Allergies, g.DietaryRequirements).Scan(&guestID); err != nil {
			return nil, fmt.Errorf("insert guest: %w", err)
		}
		guests = append(guests, models.Guest{
			ID:                  guestID,
			ItineraryID:         itinID,
			Name:                g.Name,
			Allergies:           g.Allergies,
			DietaryRequirements: g.DietaryRequirements,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &models.Itinerary{
		ID:        itinID,
		GroupName: req.GroupName,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Stops:     stops,
		Guests:    guests,
	}, nil
}

func (s *PostgresStore) ListItineraries(ctx context.Context) ([]models.Itinerary, error) {
	query := `SELECT id, group_name, created_at, updated_at FROM itineraries ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query itineraries: %w", err)
	}
	defer rows.Close()

	var result []models.Itinerary
	for rows.Next() {
		var it models.Itinerary
		if err := rows.Scan(&it.ID, &it.GroupName, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan itinerary: %w", err)
		}
		result = append(result, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	if result == nil {
		result = []models.Itinerary{}
	}

	// Populate stops and guests for each itinerary
	for i := range result {
		stops, err := s.getStopsForItinerary(ctx, result[i].ID)
		if err != nil {
			return nil, err
		}
		result[i].Stops = stops

		guests, err := s.getGuestsForItinerary(ctx, result[i].ID)
		if err != nil {
			return nil, err
		}
		result[i].Guests = guests
	}

	return result, nil
}

func (s *PostgresStore) GetItinerary(ctx context.Context, id string) (*models.Itinerary, error) {
	var it models.Itinerary
	query := `SELECT id, group_name, created_at, updated_at FROM itineraries WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&it.ID, &it.GroupName, &it.CreatedAt, &it.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query itinerary by id: %w", err)
	}

	stops, err := s.getStopsForItinerary(ctx, id)
	if err != nil {
		return nil, err
	}
	it.Stops = stops

	guests, err := s.getGuestsForItinerary(ctx, id)
	if err != nil {
		return nil, err
	}
	it.Guests = guests

	return &it, nil
}

func (s *PostgresStore) UpdateGuests(ctx context.Context, itineraryID string, guests []models.GuestInput) (*models.Itinerary, error) {
	if len(guests) == 0 {
		return nil, ErrEmptyGuests
	}

	// Verify itinerary exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM itineraries WHERE id = $1)`
	if err := s.db.QueryRowContext(ctx, checkQuery, itineraryID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check itinerary: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Remove old guests and insert new ones
	if _, err := tx.ExecContext(ctx, `DELETE FROM guests WHERE itinerary_id = $1`, itineraryID); err != nil {
		return nil, fmt.Errorf("delete old guests: %w", err)
	}

	for _, g := range guests {
		insertGuest := `
			INSERT INTO guests (itinerary_id, name, allergies, dietary_requirements)
			VALUES ($1, $2, $3, $4)`
		if _, err := tx.ExecContext(ctx, insertGuest, itineraryID, g.Name, g.Allergies, g.DietaryRequirements); err != nil {
			return nil, fmt.Errorf("insert updated guest: %w", err)
		}
	}

	// Update updated_at timestamp on the itinerary
	if _, err := tx.ExecContext(ctx, `UPDATE itineraries SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`, itineraryID); err != nil {
		return nil, fmt.Errorf("update itinerary timestamp: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetItinerary(ctx, itineraryID)
}

func (s *PostgresStore) getStopsForItinerary(ctx context.Context, itineraryID string) ([]models.Stop, error) {
	query := `
		SELECT id, itinerary_id, lodge_name, chef_phone, TO_CHAR(check_in_date, 'YYYY-MM-DD'), notification_sent, chef_confirmed
		FROM itinerary_stops
		WHERE itinerary_id = $1
		ORDER BY check_in_date ASC, id ASC`
	rows, err := s.db.QueryContext(ctx, query, itineraryID)
	if err != nil {
		return nil, fmt.Errorf("query stops: %w", err)
	}
	defer rows.Close()

	var stops []models.Stop
	for rows.Next() {
		var st models.Stop
		if err := rows.Scan(&st.ID, &st.ItineraryID, &st.LodgeName, &st.ChefPhone, &st.CheckInDate, &st.NotificationSent, &st.ChefConfirmed); err != nil {
			return nil, fmt.Errorf("scan stop: %w", err)
		}
		stops = append(stops, st)
	}
	if stops == nil {
		stops = []models.Stop{}
	}
	return stops, nil
}

func (s *PostgresStore) getGuestsForItinerary(ctx context.Context, itineraryID string) ([]models.Guest, error) {
	query := `
		SELECT id, itinerary_id, name, allergies, COALESCE(dietary_requirements, '')
		FROM guests
		WHERE itinerary_id = $1
		ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, query, itineraryID)
	if err != nil {
		return nil, fmt.Errorf("query guests: %w", err)
	}
	defer rows.Close()

	var guests []models.Guest
	for rows.Next() {
		var g models.Guest
		if err := rows.Scan(&g.ID, &g.ItineraryID, &g.Name, &g.Allergies, &g.DietaryRequirements); err != nil {
			return nil, fmt.Errorf("scan guest: %w", err)
		}
		guests = append(guests, g)
	}
	if guests == nil {
		guests = []models.Guest{}
	}
	return guests, nil
}

// GetStopByItineraryAndPhone authenticates a chef and returns the itinerary stop.
func (s *PostgresStore) GetStopByItineraryAndPhone(ctx context.Context, itineraryID, chefPhone string) (*USSDStop, error) {
	query := `
		SELECT i.id, i.group_name, st.lodge_name, st.chef_phone,
		       TO_CHAR(st.check_in_date, 'YYYY-MM-DD'), COUNT(g.id), st.chef_confirmed
		FROM itinerary_stops st
		JOIN itineraries i ON i.id = st.itinerary_id
		LEFT JOIN guests g ON g.itinerary_id = i.id
		WHERE st.itinerary_id = $1 AND st.chef_phone = $2
		GROUP BY i.id, i.group_name, st.id, st.lodge_name, st.chef_phone,
		         st.check_in_date, st.chef_confirmed`
	var stop USSDStop
	if err := s.db.QueryRowContext(ctx, query, itineraryID, chefPhone).Scan(
		&stop.ItineraryID, &stop.GroupName, &stop.LodgeName, &stop.ChefPhone,
		&stop.CheckInDate, &stop.GuestCount, &stop.ChefConfirmed,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query USSD stop: %w", err)
	}
	return &stop, nil
}

// GetGuestsByItinerary returns guest manifest data for an authenticated USSD session.
func (s *PostgresStore) GetGuestsByItinerary(ctx context.Context, itineraryID string) ([]USSDGuest, error) {
	query := `SELECT name, allergies, COALESCE(dietary_requirements, '') FROM guests WHERE itinerary_id = $1 ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, query, itineraryID)
	if err != nil {
		return nil, fmt.Errorf("query USSD guests: %w", err)
	}
	defer rows.Close()

	var guests []USSDGuest
	for rows.Next() {
		var guest USSDGuest
		if err := rows.Scan(&guest.Name, &guest.Allergies, &guest.DietaryRequirements); err != nil {
			return nil, fmt.Errorf("scan USSD guest: %w", err)
		}
		guests = append(guests, guest)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate USSD guests: %w", err)
	}
	return guests, nil
}

// ConfirmManifest records receipt only for the stop authenticated by itinerary and phone.
func (s *PostgresStore) ConfirmManifest(ctx context.Context, itineraryID, chefPhone string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE itinerary_stops SET chef_confirmed = TRUE WHERE itinerary_id = $1 AND chef_phone = $2`,
		itineraryID, chefPhone)
	if err != nil {
		return fmt.Errorf("confirm manifest: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check manifest confirmation: %w", err)
	}
	if updated == 0 {
		return ErrNotFound
	}
	return nil
}

// QueryPendingStops returns unnotified stops scheduled from today through the given horizon.
func (s *PostgresStore) QueryPendingStops(ctx context.Context, within time.Duration) ([]PendingStop, error) {
	cutoff := time.Now().Add(within).Format("2006-01-02")
	query := `
		SELECT st.id, i.id, i.group_name, st.chef_phone, st.lodge_name,
		       TO_CHAR(st.check_in_date, 'YYYY-MM-DD'), COUNT(g.id)
		FROM itinerary_stops st
		JOIN itineraries i ON i.id = st.itinerary_id
		LEFT JOIN guests g ON g.itinerary_id = i.id
		WHERE st.notification_sent = FALSE
		  AND st.check_in_date >= CURRENT_DATE
		  AND st.check_in_date <= $1::date
		GROUP BY st.id, i.id, i.group_name, st.chef_phone, st.lodge_name, st.check_in_date
		ORDER BY st.check_in_date ASC, st.id ASC`
	rows, err := s.db.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query pending stops: %w", err)
	}
	defer rows.Close()

	var stops []PendingStop
	for rows.Next() {
		var stop PendingStop
		if err := rows.Scan(&stop.StopID, &stop.ItineraryID, &stop.GroupName, &stop.ChefPhone,
			&stop.LodgeName, &stop.CheckInDate, &stop.GuestCount); err != nil {
			return nil, fmt.Errorf("scan pending stop: %w", err)
		}
		stops = append(stops, stop)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending stops: %w", err)
	}
	return stops, nil
}

// MarkNotificationSent records successful notification delivery for a stop.
func (s *PostgresStore) MarkNotificationSent(ctx context.Context, stopID int) error {
	result, err := s.db.ExecContext(ctx, `UPDATE itinerary_stops SET notification_sent = TRUE WHERE id = $1`, stopID)
	if err != nil {
		return fmt.Errorf("mark notification sent: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check notification update: %w", err)
	}
	if updated == 0 {
		return ErrNotFound
	}
	return nil
}

// MemoryStore implements an in-memory Store for testing and offline/local execution.
type MemoryStore struct {
	mu          sync.RWMutex
	itineraries map[string]*models.Itinerary
	nextID      int
	nextStopID  int
}

// NewMemoryStore creates an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		itineraries: make(map[string]*models.Itinerary),
		nextID:      1,
		nextStopID:  1,
	}
}

func (m *MemoryStore) Ping(ctx context.Context) error {
	return nil
}

func (m *MemoryStore) Close() error {
	return nil
}

func (m *MemoryStore) CreateItinerary(ctx context.Context, req models.CreateItineraryRequest) (*models.Itinerary, error) {
	if strings.TrimSpace(req.GroupName) == "" {
		return nil, ErrEmptyGroupName
	}
	if len(req.Stops) == 0 {
		return nil, ErrEmptyStops
	}
	if len(req.Guests) == 0 {
		return nil, ErrEmptyGuests
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	itinID := strings.TrimSpace(req.ID)
	if itinID == "" {
		itinID = fmt.Sprintf("ITIN-%04d", m.nextID)
		m.nextID++
	}

	now := time.Now().UTC()
	stops := make([]models.Stop, len(req.Stops))
	for i, s := range req.Stops {
		stops[i] = models.Stop{
			ID:               m.nextStopID,
			ItineraryID:      itinID,
			LodgeName:        s.LodgeName,
			ChefPhone:        s.ChefPhone,
			CheckInDate:      s.CheckInDate,
			NotificationSent: false,
			ChefConfirmed:    false,
		}
		m.nextStopID++
	}

	guests := make([]models.Guest, len(req.Guests))
	for i, g := range req.Guests {
		guests[i] = models.Guest{
			ID:                  i + 1,
			ItineraryID:         itinID,
			Name:                g.Name,
			Allergies:           g.Allergies,
			DietaryRequirements: g.DietaryRequirements,
		}
	}

	it := &models.Itinerary{
		ID:        itinID,
		GroupName: req.GroupName,
		CreatedAt: now,
		UpdatedAt: now,
		Stops:     stops,
		Guests:    guests,
	}

	m.itineraries[itinID] = it
	return it, nil
}

func (m *MemoryStore) ListItineraries(ctx context.Context) ([]models.Itinerary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]models.Itinerary, 0, len(m.itineraries))
	for _, it := range m.itineraries {
		res = append(res, *it)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt.After(res[j].CreatedAt)
	})
	return res, nil
}

func (m *MemoryStore) GetItinerary(ctx context.Context, id string) (*models.Itinerary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	it, ok := m.itineraries[id]
	if !ok {
		return nil, ErrNotFound
	}
	copyIt := *it
	return &copyIt, nil
}

func (m *MemoryStore) UpdateGuests(ctx context.Context, itineraryID string, guests []models.GuestInput) (*models.Itinerary, error) {
	if len(guests) == 0 {
		return nil, ErrEmptyGuests
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	it, ok := m.itineraries[itineraryID]
	if !ok {
		return nil, ErrNotFound
	}

	updatedGuests := make([]models.Guest, len(guests))
	for i, g := range guests {
		updatedGuests[i] = models.Guest{
			ID:                  i + 1,
			ItineraryID:         itineraryID,
			Name:                g.Name,
			Allergies:           g.Allergies,
			DietaryRequirements: g.DietaryRequirements,
		}
	}

	it.Guests = updatedGuests
	it.UpdatedAt = time.Now().UTC()
	copyIt := *it
	return &copyIt, nil
}

func (m *MemoryStore) GetStopByItineraryAndPhone(_ context.Context, itineraryID, chefPhone string) (*USSDStop, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	it, ok := m.itineraries[itineraryID]
	if !ok {
		return nil, ErrNotFound
	}
	for _, stop := range it.Stops {
		if stop.ChefPhone == chefPhone {
			return &USSDStop{
				ItineraryID:   it.ID,
				GroupName:     it.GroupName,
				LodgeName:     stop.LodgeName,
				ChefPhone:     stop.ChefPhone,
				CheckInDate:   stop.CheckInDate,
				GuestCount:    len(it.Guests),
				ChefConfirmed: stop.ChefConfirmed,
			}, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MemoryStore) GetGuestsByItinerary(_ context.Context, itineraryID string) ([]USSDGuest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	it, ok := m.itineraries[itineraryID]
	if !ok {
		return nil, ErrNotFound
	}
	guests := make([]USSDGuest, len(it.Guests))
	for i, guest := range it.Guests {
		guests[i] = USSDGuest{
			Name:                guest.Name,
			Allergies:           guest.Allergies,
			DietaryRequirements: guest.DietaryRequirements,
		}
	}
	return guests, nil
}

func (m *MemoryStore) ConfirmManifest(_ context.Context, itineraryID, chefPhone string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	it, ok := m.itineraries[itineraryID]
	if !ok {
		return ErrNotFound
	}
	for i := range it.Stops {
		if it.Stops[i].ChefPhone == chefPhone {
			it.Stops[i].ChefConfirmed = true
			return nil
		}
	}
	return ErrNotFound
}

func (m *MemoryStore) QueryPendingStops(_ context.Context, within time.Duration) ([]PendingStop, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	today := now.Format("2006-01-02")
	cutoff := now.Add(within).Format("2006-01-02")
	var pending []PendingStop
	for _, it := range m.itineraries {
		for _, stop := range it.Stops {
			if stop.NotificationSent || stop.CheckInDate < today || stop.CheckInDate > cutoff {
				continue
			}
			pending = append(pending, PendingStop{
				StopID: stop.ID, ItineraryID: it.ID, GroupName: it.GroupName,
				ChefPhone: stop.ChefPhone, LodgeName: stop.LodgeName,
				CheckInDate: stop.CheckInDate, GuestCount: len(it.Guests),
			})
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		if pending[i].CheckInDate == pending[j].CheckInDate {
			return pending[i].StopID < pending[j].StopID
		}
		return pending[i].CheckInDate < pending[j].CheckInDate
	})
	return pending, nil
}

func (m *MemoryStore) MarkNotificationSent(_ context.Context, stopID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, it := range m.itineraries {
		for i := range it.Stops {
			if it.Stops[i].ID == stopID {
				it.Stops[i].NotificationSent = true
				return nil
			}
		}
	}
	return ErrNotFound
}
