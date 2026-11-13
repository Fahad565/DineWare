import { CreateItineraryRequest, Itinerary, UpdateGuestsRequest } from '../types/itinerary';
import { INITIAL_MOCK_ITINERARIES } from './mockData';

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const LOCAL_STORAGE_KEY = 'dineware_itineraries_cache';

// Helper to interact with local storage fallback
function getLocalFallback(): Itinerary[] {
  if (typeof window === 'undefined') {
    return INITIAL_MOCK_ITINERARIES;
  }
  try {
    const raw = localStorage.getItem(LOCAL_STORAGE_KEY);
    if (!raw) {
      localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(INITIAL_MOCK_ITINERARIES));
      return INITIAL_MOCK_ITINERARIES;
    }
    return JSON.parse(raw);
  } catch {
    return INITIAL_MOCK_ITINERARIES;
  }
}

function saveLocalFallback(data: Itinerary[]): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(data));
  } catch (e) {
    console.warn('Failed to save to localStorage:', e);
  }
}

export async function checkBackendHealth(): Promise<{ online: boolean; message?: string }> {
  try {
    const res = await fetch(`${BASE_URL}/health`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
      cache: 'no-store',
    });
    if (res.ok) {
      return { online: true };
    }
    return { online: false, message: `Status ${res.status}` };
  } catch {
    return { online: false, message: 'Unreachable' };
  }
}

export async function fetchItineraries(): Promise<Itinerary[]> {
  try {
    const res = await fetch(`${BASE_URL}/api/v1/itineraries`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
      cache: 'no-store',
    });

    if (res.ok) {
      const data: Itinerary[] = await res.json();
      if (Array.isArray(data) && data.length > 0) {
        saveLocalFallback(data);
        return data;
      }
    }
  } catch (e) {
    console.info('Backend unreachable, using local persistent fallback store:', e);
  }

  return getLocalFallback();
}

export async function fetchItinerary(id: string): Promise<Itinerary | null> {
  try {
    const res = await fetch(`${BASE_URL}/api/v1/itineraries/${encodeURIComponent(id)}`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
      cache: 'no-store',
    });

    if (res.ok) {
      const data: Itinerary = await res.json();
      return data;
    }
  } catch (e) {
    console.info(`Backend fetch for itinerary ${id} failed, checking local store:`, e);
  }

  const localList = getLocalFallback();
  const match = localList.find((it) => it.id.toLowerCase() === id.toLowerCase());
  return match || null;
}

export async function createItinerary(req: CreateItineraryRequest): Promise<Itinerary> {
  const generatedId = req.id && req.id.trim() !== '' 
    ? req.id.trim() 
    : `ITIN-2026-${Math.floor(100 + Math.random() * 900)}`;

  const payload: CreateItineraryRequest = {
    ...req,
    id: generatedId,
  };

  try {
    const res = await fetch(`${BASE_URL}/api/v1/itineraries`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify(payload),
    });

    if (res.ok) {
      const created: Itinerary = await res.json();
      const local = getLocalFallback();
      saveLocalFallback([created, ...local.filter((i) => i.id !== created.id)]);
      return created;
    }
  } catch (e) {
    console.warn('Backend POST failed, storing itinerary in local cache:', e);
  }

  // Local fallback creation
  const now = new Date().toISOString();
  const newItinerary: Itinerary = {
    id: generatedId,
    group_name: req.group_name,
    created_at: now,
    updated_at: now,
    stops: req.stops.map((s, idx) => ({
      id: idx + 1,
      itinerary_id: generatedId,
      lodge_name: s.lodge_name,
      chef_phone: s.chef_phone,
      check_in_date: s.check_in_date,
      notification_sent: false,
      chef_confirmed: false,
    })),
    guests: req.guests.map((g, idx) => ({
      id: idx + 1,
      itinerary_id: generatedId,
      name: g.name,
      allergies: g.allergies,
      dietary_requirements: g.dietary_requirements || 'Standard',
    })),
  };

  const local = getLocalFallback();
  saveLocalFallback([newItinerary, ...local]);
  return newItinerary;
}

export async function updateGuests(id: string, req: UpdateGuestsRequest): Promise<Itinerary> {
  try {
    const res = await fetch(`${BASE_URL}/api/v1/itineraries/${encodeURIComponent(id)}/guests`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify(req),
    });

    if (res.ok) {
      const updated: Itinerary = await res.json();
      const local = getLocalFallback();
      saveLocalFallback(local.map((item) => (item.id === updated.id ? updated : item)));
      return updated;
    }
  } catch (e) {
    console.warn(`Backend PATCH failed for ${id}, updating local cache:`, e);
  }

  // Local fallback update
  const local = getLocalFallback();
  const index = local.findIndex((i) => i.id.toLowerCase() === id.toLowerCase());
  if (index === -1) {
    throw new Error('Itinerary not found in local cache');
  }

  const existing = local[index];
  const updatedItinerary: Itinerary = {
    ...existing,
    updated_at: new Date().toISOString(),
    guests: req.guests.map((g, idx) => ({
      id: idx + 1,
      itinerary_id: existing.id,
      name: g.name,
      allergies: g.allergies,
      dietary_requirements: g.dietary_requirements || 'Standard',
    })),
  };

  local[index] = updatedItinerary;
  saveLocalFallback(local);
  return updatedItinerary;
}

export async function confirmManifestForLodge(itineraryId: string, stopId?: number): Promise<Itinerary | null> {
  const local = getLocalFallback();
  const index = local.findIndex((i) => i.id.toLowerCase() === itineraryId.toLowerCase());
  if (index === -1) return null;

  const target = local[index];
  const updatedStops = target.stops.map((stop) => {
    if (!stopId || stop.id === stopId || target.stops.length === 1) {
      return {
        ...stop,
        notification_sent: true,
        chef_confirmed: true,
      };
    }
    return stop;
  });

  const updated: Itinerary = {
    ...target,
    updated_at: new Date().toISOString(),
    stops: updatedStops,
  };

  local[index] = updated;
  saveLocalFallback(local);
  return updated;
}
