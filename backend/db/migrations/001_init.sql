-- 001_init.sql: Define core tables for DineWare backend
-- Tables: itineraries, itinerary_stops, guests (storing protected dietary & allergy specs)

CREATE TABLE IF NOT EXISTS itineraries (
    id VARCHAR(50) PRIMARY KEY,
    group_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS itinerary_stops (
    id SERIAL PRIMARY KEY,
    itinerary_id VARCHAR(50) NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    lodge_name VARCHAR(100) NOT NULL,
    chef_phone VARCHAR(20) NOT NULL,
    check_in_date DATE NOT NULL,
    notification_sent BOOLEAN DEFAULT FALSE,
    chef_confirmed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS guests (
    id SERIAL PRIMARY KEY,
    itinerary_id VARCHAR(50) NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    allergies TEXT NOT NULL,
    dietary_requirements TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indices for performance on lookups and scheduling
CREATE INDEX IF NOT EXISTS idx_itinerary_stops_itinerary_id ON itinerary_stops(itinerary_id);
CREATE INDEX IF NOT EXISTS idx_itinerary_stops_check_in_date ON itinerary_stops(check_in_date);
CREATE INDEX IF NOT EXISTS idx_itinerary_stops_chef_phone ON itinerary_stops(chef_phone);
CREATE INDEX IF NOT EXISTS idx_guests_itinerary_id ON guests(itinerary_id);
