'use client';

import React, { useEffect, useState } from 'react';
import { AlertCircle, Building2, Calendar, Check, Plus, Trash2, User, X } from 'lucide-react';
import {
  GuestInput,
  Itinerary,
  StopInput,
  UpdateItineraryRequest,
  UpdateStopInput,
} from '../types/itinerary';
import { updateItinerary } from '../lib/api';

interface EditItineraryModalProps {
  isOpen: boolean;
  onClose: () => void;
  itinerary: Itinerary | null;
  onUpdated: (updated: Itinerary) => void;
}

export function EditItineraryModal({
  isOpen,
  onClose,
  itinerary,
  onUpdated,
}: EditItineraryModalProps) {
  const [groupName, setGroupName] = useState('');
  const [stops, setStops] = useState<UpdateStopInput[]>([]);
  const [guests, setGuests] = useState<GuestInput[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!itinerary) return;
    setGroupName(itinerary.group_name);
    setStops(itinerary.stops.map((stop) => ({
      id: stop.id,
      lodge_name: stop.lodge_name,
      chef_phone: stop.chef_phone,
      check_in_date: stop.check_in_date,
    })));
    setGuests(itinerary.guests.map((guest) => ({
      name: guest.name,
      allergies: guest.allergies,
      dietary_requirements: guest.dietary_requirements || '',
    })));
    setError(null);
  }, [itinerary]);

  if (!isOpen || !itinerary) return null;

  const updateStop = (index: number, field: keyof StopInput, value: string) => {
    setStops((current) => current.map((stop, i) => (
      i === index ? { ...stop, [field]: value } : stop
    )));
  };

  const updateGuest = (index: number, field: keyof GuestInput, value: string) => {
    setGuests((current) => current.map((guest, i) => (
      i === index ? { ...guest, [field]: value } : guest
    )));
  };

  const addStop = () => {
    setStops((current) => [...current, {
      lodge_name: '',
      chef_phone: '',
      check_in_date: new Date(Date.now() + 3 * 86400000).toISOString().slice(0, 10),
    }]);
  };

  const addGuest = () => {
    setGuests((current) => [...current, { name: '', allergies: '', dietary_requirements: '' }]);
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    if (!groupName.trim()) {
      setError('Group name is required.');
      return;
    }
    if (stops.length === 0 || stops.some((stop) => (
      !stop.lodge_name.trim() || !stop.chef_phone.trim() || !stop.check_in_date
    ))) {
      setError('Add at least one complete lodge stop, including its check-in date and chef phone.');
      return;
    }
    if (guests.length === 0 || guests.some((guest) => !guest.name.trim())) {
      setError('Add at least one guest and provide a name for every guest.');
      return;
    }

    setLoading(true);
    try {
      const request: UpdateItineraryRequest = {
        group_name: groupName.trim(),
        stops,
        guests,
      };
      const updated = await updateItinerary(itinerary.id, request);
      onUpdated(updated);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update itinerary.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-stone-900/60 p-4 backdrop-blur-sm">
      <div className="flex max-h-[92vh] w-full max-w-4xl flex-col overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-2xl">
        <div className="flex items-center justify-between border-b border-stone-800 bg-stone-900 px-6 py-4 text-stone-100">
          <div>
            <h2 className="font-display text-lg font-bold text-white">Edit trip</h2>
            <p className="text-xs text-stone-400">{itinerary.id}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1.5 text-stone-400 transition-colors hover:bg-stone-800 hover:text-white"
            aria-label="Close edit trip"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 space-y-5 overflow-y-auto p-5 text-sm sm:p-6">
          {error && (
            <div role="alert" className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 p-3 text-red-700">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <label className="block">
            <span className="mb-1 block text-xs font-semibold text-stone-700">Group name</span>
            <input
              required
              value={groupName}
              onChange={(event) => setGroupName(event.target.value)}
              className="w-full rounded-lg border border-stone-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-500"
            />
          </label>

          <section className="space-y-3 rounded-xl border border-stone-200 bg-stone-50 p-4">
            <div className="flex items-center justify-between gap-3">
              <h3 className="flex items-center gap-2 font-semibold text-stone-900">
                <Building2 className="h-4 w-4 text-amber-700" />
                Lodge stops
              </h3>
              <button
                type="button"
                onClick={addStop}
                className="inline-flex items-center gap-1 rounded-lg bg-amber-100 px-3 py-1.5 text-xs font-medium text-amber-900 hover:bg-amber-200"
              >
                <Plus className="h-3.5 w-3.5" />
                Add stop
              </button>
            </div>
            {stops.map((stop, index) => (
              <div
                key={stop.id ?? `new-stop-${index}`}
                className="grid grid-cols-1 items-end gap-3 rounded-lg border border-stone-200 bg-white p-3 sm:grid-cols-12"
              >
                <label className="sm:col-span-4">
                  <span className="mb-1 block text-[11px] font-medium text-stone-600">Lodge name</span>
                  <input
                    required
                    value={stop.lodge_name}
                    onChange={(event) => updateStop(index, 'lodge_name', event.target.value)}
                    className="w-full rounded-md border border-stone-300 px-2.5 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500"
                  />
                </label>
                <label className="sm:col-span-3">
                  <span className="mb-1 block text-[11px] font-medium text-stone-600">Chef phone</span>
                  <input
                    required
                    type="tel"
                    value={stop.chef_phone}
                    onChange={(event) => updateStop(index, 'chef_phone', event.target.value)}
                    className="w-full rounded-md border border-stone-300 px-2.5 py-2 font-mono text-xs focus:outline-none focus:ring-2 focus:ring-amber-500"
                  />
                </label>
                <label className="sm:col-span-4">
                  <span className="mb-1 flex items-center gap-1 text-[11px] font-medium text-stone-600">
                    <Calendar className="h-3 w-3" />
                    Check-in date
                  </span>
                  <input
                    required
                    type="date"
                    value={stop.check_in_date}
                    onChange={(event) => updateStop(index, 'check_in_date', event.target.value)}
                    className="w-full rounded-md border border-stone-300 px-2.5 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500"
                  />
                </label>
                <button
                  type="button"
                  disabled={stops.length <= 1}
                  onClick={() => setStops((current) => current.filter((_, i) => i !== index))}
                  className="justify-self-end rounded p-2 text-stone-400 hover:bg-red-50 hover:text-red-700 disabled:opacity-30 sm:col-span-1"
                  title="Remove lodge stop"
                  aria-label={`Remove lodge stop ${index + 1}`}
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
          </section>

          <section className="space-y-3 rounded-xl border border-stone-200 bg-stone-50 p-4">
            <div className="flex items-center justify-between gap-3">
              <h3 className="flex items-center gap-2 font-semibold text-stone-900">
                <User className="h-4 w-4 text-amber-700" />
                Guests and meal requirements
              </h3>
              <button
                type="button"
                onClick={addGuest}
                className="inline-flex items-center gap-1 rounded-lg bg-amber-100 px-3 py-1.5 text-xs font-medium text-amber-900 hover:bg-amber-200"
              >
                <Plus className="h-3.5 w-3.5" />
                Add guest
              </button>
            </div>
            {guests.map((guest, index) => (
              <div key={`guest-${index}`} className="grid grid-cols-1 items-end gap-3 rounded-lg border border-stone-200 bg-white p-3 sm:grid-cols-12">
                <label className="sm:col-span-3">
                  <span className="mb-1 block text-[11px] font-medium text-stone-600">Guest name</span>
                  <input
                    required
                    value={guest.name}
                    onChange={(event) => updateGuest(index, 'name', event.target.value)}
                    className="w-full rounded-md border border-stone-300 px-2.5 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500"
                  />
                </label>
                <label className="sm:col-span-4">
                  <span className="mb-1 block text-[11px] font-medium text-stone-600">Allergies</span>
                  <input
                    value={guest.allergies}
                    onChange={(event) => updateGuest(index, 'allergies', event.target.value)}
                    className="w-full rounded-md border border-stone-300 px-2.5 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500"
                  />
                </label>
                <label className="sm:col-span-4">
                  <span className="mb-1 block text-[11px] font-medium text-stone-600">Dietary requirements</span>
                  <input
                    value={guest.dietary_requirements || ''}
                    onChange={(event) => updateGuest(index, 'dietary_requirements', event.target.value)}
                    className="w-full rounded-md border border-stone-300 px-2.5 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500"
                  />
                </label>
                <button
                  type="button"
                  disabled={guests.length <= 1}
                  onClick={() => setGuests((current) => current.filter((_, i) => i !== index))}
                  className="justify-self-end rounded p-2 text-stone-400 hover:bg-red-50 hover:text-red-700 disabled:opacity-30 sm:col-span-1"
                  title="Remove guest"
                  aria-label={`Remove guest ${index + 1}`}
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
          </section>

          <div className="flex items-center justify-end gap-3 border-t border-stone-200 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-stone-300 px-4 py-2 font-medium text-stone-700 hover:bg-stone-100"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="inline-flex items-center gap-2 rounded-lg bg-amber-700 px-5 py-2 font-semibold text-white hover:bg-amber-800 disabled:opacity-50"
            >
              <Check className="h-4 w-4" />
              {loading ? 'Saving…' : 'Save trip'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
