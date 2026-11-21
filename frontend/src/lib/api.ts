import { CreateItineraryRequest, Itinerary, UpdateGuestsRequest } from '../types/itinerary';

const API_BASE_URL = '/api/backend';

async function readError(res: Response): Promise<Error> {
  const body = await res.text();
  return new Error(body || `Backend request failed with status ${res.status}`);
}

export async function checkBackendHealth(): Promise<{ online: boolean; message?: string }> {
  try {
    const res = await fetch(`${API_BASE_URL}/health`, {
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
  const res = await fetch(`${API_BASE_URL}/api/v1/itineraries`, {
    method: 'GET',
    headers: { Accept: 'application/json' },
    cache: 'no-store',
  });
  if (!res.ok) {
    throw await readError(res);
  }

  const data: unknown = await res.json();
  if (!Array.isArray(data)) {
    throw new Error('Backend returned an invalid itinerary list');
  }
  return data as Itinerary[];
}

export async function fetchItinerary(id: string): Promise<Itinerary | null> {
  const res = await fetch(`${API_BASE_URL}/api/v1/itineraries/${encodeURIComponent(id)}`, {
    method: 'GET',
    headers: { Accept: 'application/json' },
    cache: 'no-store',
  });
  if (res.status === 404) {
    return null;
  }
  if (!res.ok) {
    throw await readError(res);
  }
  return res.json();
}

export async function createItinerary(req: CreateItineraryRequest): Promise<Itinerary> {
  const payload: CreateItineraryRequest = {
    ...req,
    id: req.id?.trim() || undefined,
  };
  const res = await fetch(`${API_BASE_URL}/api/v1/itineraries`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    throw await readError(res);
  }
  return res.json();
}

export async function updateGuests(id: string, req: UpdateGuestsRequest): Promise<Itinerary> {
  const res = await fetch(`${API_BASE_URL}/api/v1/itineraries/${encodeURIComponent(id)}/guests`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    throw await readError(res);
  }
  return res.json();
}
