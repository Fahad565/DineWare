'use client';

import React, { useState } from 'react';
import { X, Plus, Trash2, User, AlertCircle, RefreshCw, Zap, ShieldAlert } from 'lucide-react';
import { GuestInput, Itinerary } from '../types/itinerary';
import { updateGuests } from '../lib/api';

interface EditGuestsModalProps {
  isOpen: boolean;
  onClose: () => void;
  itinerary: Itinerary | null;
  onUpdated: (updated: Itinerary) => void;
}

const COMMON_ALLERGIES = [
  'Peanuts (Severe Anaphylaxis)',
  'Tree Nuts',
  'Shellfish',
  'Gluten / Celiac',
  'Dairy / Lactose',
  'Soy',
  'Egg',
];

export function EditGuestsModal({ isOpen, onClose, itinerary, onUpdated }: EditGuestsModalProps) {
  const [guests, setGuests] = useState<GuestInput[]>(
    itinerary
      ? itinerary.guests.map((g) => ({
          name: g.name,
          allergies: g.allergies,
          dietary_requirements: g.dietary_requirements || 'Standard',
        }))
      : []
  );

  // Sync state if itinerary changes
  React.useEffect(() => {
    if (itinerary) {
      setGuests(
        itinerary.guests.map((g) => ({
          name: g.name,
          allergies: g.allergies,
          dietary_requirements: g.dietary_requirements || 'Standard',
        }))
      );
    }
  }, [itinerary]);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [propagationNotice, setPropagationNotice] = useState<string | null>(null);

  if (!isOpen || !itinerary) return null;

  const handleAddGuest = () => {
    setGuests([...guests, { name: '', allergies: 'None', dietary_requirements: 'Standard Safari Menu' }]);
  };

  const handleRemoveGuest = (index: number) => {
    if (guests.length <= 1) {
      setError('An itinerary must contain at least one guest.');
      return;
    }
    setError(null);
    setGuests(guests.filter((_, i) => i !== index));
  };

  const handleGuestChange = (index: number, field: keyof GuestInput, value: string) => {
    const updated = [...guests];
    updated[index] = { ...updated[index], [field]: value };
    setGuests(updated);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (guests.some((g) => !g.name.trim())) {
      setError('All guests must have a valid name.');
      return;
    }

    setLoading(true);
    try {
      const updated = await updateGuests(itinerary.id, { guests });
      
      const downstreamLodgeNames = itinerary.stops.map((s) => s.lodge_name).join(' → ');
      setPropagationNotice(`Real-time downstream update dispatched! Propagated to: ${downstreamLodgeNames}`);
      
      setTimeout(() => {
        onUpdated(updated);
        onClose();
        setPropagationNotice(null);
      }, 1200);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to update guest manifest';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-stone-900/60 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl border border-stone-200 max-w-3xl w-full max-h-[92vh] flex flex-col overflow-hidden animate-in fade-in duration-200">
        {/* Header */}
        <div className="px-6 py-4 bg-stone-900 text-stone-100 flex items-center justify-between border-b border-stone-800">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 rounded-lg bg-amber-600 flex items-center justify-center text-white">
              <Zap className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white flex items-center space-x-2">
                <span>Mid-Trip Manifest &amp; Allergy Propagation</span>
              </h2>
              <p className="text-xs text-stone-400">
                Itinerary: <span className="font-mono text-amber-300 font-semibold">{itinerary.id}</span> ({itinerary.group_name})
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-stone-400 hover:text-stone-200 p-1.5 rounded-lg hover:bg-stone-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Real-time propagation banner */}
        <div className="bg-gradient-to-r from-amber-500/10 via-amber-500/5 to-transparent border-b border-amber-200 px-6 py-3 flex items-start space-x-3 text-xs text-amber-900">
          <ShieldAlert className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
          <div>
            <span className="font-bold">Real-Time Downstream Synchronization:</span> Any modifications to guest counts, anaphylactic allergies, or dietary specs will immediately update Neon PostgreSQL and synchronize with downstream lodge kitchens via Africa’s Talking USSD and 72h SMS queues.
          </div>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-5 text-sm">
          {error && (
            <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-red-700 flex items-start space-x-2">
              <AlertCircle className="w-5 h-5 shrink-0 mt-0.5 text-red-500" />
              <span>{error}</span>
            </div>
          )}

          {propagationNotice && (
            <div className="p-3 rounded-lg bg-emerald-50 border border-emerald-300 text-emerald-800 flex items-center space-x-2 animate-pulse">
              <RefreshCw className="w-4 h-4 animate-spin text-emerald-600" />
              <span className="font-medium">{propagationNotice}</span>
            </div>
          )}

          {/* Stops impacted info */}
          <div className="bg-stone-50 p-3.5 rounded-xl border border-stone-200 text-xs text-stone-700">
            <span className="font-bold text-stone-800">Downstream Stops Being Updated:</span>
            <div className="flex flex-wrap gap-2 mt-2">
              {itinerary.stops.map((stop, idx) => (
                <div
                  key={idx}
                  className="px-2.5 py-1 rounded-md bg-white border border-stone-200 flex items-center space-x-1.5 shadow-sm"
                >
                  <span className="w-2 h-2 rounded-full bg-amber-500"></span>
                  <span className="font-medium text-stone-900">{stop.lodge_name}</span>
                  <span className="text-stone-400 font-mono text-[10px]">({stop.check_in_date})</span>
                </div>
              ))}
            </div>
          </div>

          {/* Guest Table */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold text-stone-900 flex items-center space-x-2">
                <User className="w-4 h-4 text-amber-600" />
                <span>Guest Count &amp; Allergy Manifest ({guests.length} Guests)</span>
              </h3>
              <button
                type="button"
                onClick={handleAddGuest}
                className="inline-flex items-center space-x-1 text-xs font-medium bg-amber-100 text-amber-800 px-3 py-1.5 rounded-lg hover:bg-amber-200 transition-colors"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add Guest Mid-Trip</span>
              </button>
            </div>

            <div className="space-y-3">
              {guests.map((guest, index) => (
                <div
                  key={index}
                  className="bg-stone-50 p-3.5 rounded-lg border border-stone-200 space-y-2"
                >
                  <div className="grid grid-cols-1 md:grid-cols-12 gap-3 items-center">
                    <div className="md:col-span-3">
                      <label className="block text-[11px] font-medium text-stone-600 mb-0.5">Guest Name</label>
                      <input
                        type="text"
                        required
                        value={guest.name}
                        onChange={(e) => handleGuestChange(index, 'name', e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs bg-white focus:ring-2 focus:ring-amber-500 focus:outline-none"
                      />
                    </div>
                    <div className="md:col-span-5">
                      <label className="block text-[11px] font-medium text-stone-600 mb-0.5">
                        Allergies <span className="text-red-500 font-semibold">(Protected Health Data)</span>
                      </label>
                      <input
                        type="text"
                        required
                        value={guest.allergies}
                        onChange={(e) => handleGuestChange(index, 'allergies', e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs bg-white focus:ring-2 focus:ring-amber-500 focus:outline-none font-medium text-stone-800"
                      />
                    </div>
                    <div className="md:col-span-3">
                      <label className="block text-[11px] font-medium text-stone-600 mb-0.5">Dietary Specs</label>
                      <input
                        type="text"
                        value={guest.dietary_requirements || ''}
                        onChange={(e) => handleGuestChange(index, 'dietary_requirements', e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs bg-white focus:ring-2 focus:ring-amber-500 focus:outline-none"
                      />
                    </div>
                    <div className="md:col-span-1 flex justify-end">
                      <button
                        type="button"
                        disabled={guests.length <= 1}
                        onClick={() => handleRemoveGuest(index)}
                        className="text-stone-400 hover:text-red-600 disabled:opacity-30 p-1.5 rounded transition-colors"
                        title="Remove guest"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>

                  {/* Allergy preset pills */}
                  <div className="flex flex-wrap items-center gap-1.5 text-[10px]">
                    <span className="text-stone-400">Append Allergy:</span>
                    {COMMON_ALLERGIES.slice(0, 4).map((a) => (
                      <button
                        key={a}
                        type="button"
                        onClick={() => {
                          const current = guest.allergies === 'None' || !guest.allergies ? '' : guest.allergies;
                          const nextVal = current ? `${current}, ${a}` : a;
                          handleGuestChange(index, 'allergies', nextVal);
                        }}
                        className="px-1.5 py-0.5 bg-white hover:bg-stone-200 text-stone-700 rounded border border-stone-200"
                      >
                        + {a}
                      </button>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Footer Submit */}
          <div className="pt-4 border-t border-stone-200 flex items-center justify-between">
            <div className="text-xs text-stone-500">
              Downstream Updated Count: <span className="font-bold text-stone-800">{guests.length} Guests</span>
            </div>
            <div className="flex items-center space-x-3">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 rounded-lg border border-stone-300 text-stone-700 hover:bg-stone-100 font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                className="px-5 py-2 rounded-lg bg-amber-600 hover:bg-amber-700 text-white font-semibold shadow-md transition-colors flex items-center space-x-2 disabled:opacity-50"
              >
                {loading ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    <span>Propagating Changes...</span>
                  </>
                ) : (
                  <>
                    <Zap className="w-4 h-4" />
                    <span>Propagate Manifest Updates</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
