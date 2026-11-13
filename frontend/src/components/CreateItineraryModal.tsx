'use client';

import React, { useState } from 'react';
import { X, Plus, Trash2, Calendar, Phone, Building2, User, AlertCircle, Sparkles } from 'lucide-react';
import { CreateItineraryRequest, GuestInput, Itinerary, StopInput } from '../types/itinerary';
import { createItinerary } from '../lib/api';

interface CreateItineraryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated: (itinerary: Itinerary) => void;
}

const COMMON_ALLERGIES = [
  'Peanuts (Severe Anaphylaxis)',
  'Tree Nuts',
  'Shellfish',
  'Gluten / Celiac',
  'Dairy / Lactose',
  'Egg',
  'None',
];

const COMMON_DIETS = [
  'Standard Safari Menu',
  'Strict Vegetarian',
  'Vegan',
  'Halal Certified',
  'Kosher',
  'Diabetic / Low Sugar',
];

export function CreateItineraryModal({ isOpen, onClose, onCreated }: CreateItineraryModalProps) {
  const [groupName, setGroupName] = useState('');
  const [customId, setCustomId] = useState('');
  const [stops, setStops] = useState<StopInput[]>([
    {
      lodge_name: 'Mara Bush Camp',
      chef_phone: '+254711223344',
      check_in_date: new Date(Date.now() + 3 * 86400000).toISOString().split('T')[0],
    },
    {
      lodge_name: 'Serengeti Serena Safari Lodge',
      chef_phone: '+255788112233',
      check_in_date: new Date(Date.now() + 6 * 86400000).toISOString().split('T')[0],
    },
  ]);
  const [guests, setGuests] = useState<GuestInput[]>([
    {
      name: 'Eleanor Vance',
      allergies: 'Peanuts (Severe Anaphylaxis)',
      dietary_requirements: 'Strict Nut-Free Prep Area',
    },
    {
      name: 'Thomas Sterling',
      allergies: 'Gluten / Celiac',
      dietary_requirements: 'Gluten-Free Menu',
    },
  ]);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleAddStop = () => {
    const lastDate = stops.length > 0 ? stops[stops.length - 1].check_in_date : new Date().toISOString().split('T')[0];
    const nextDate = new Date(new Date(lastDate).getTime() + 2 * 86400000).toISOString().split('T')[0];
    setStops([...stops, { lodge_name: '', chef_phone: '+2547', check_in_date: nextDate }]);
  };

  const handleRemoveStop = (index: number) => {
    if (stops.length <= 1) return;
    setStops(stops.filter((_, i) => i !== index));
  };

  const handleStopChange = (index: number, field: keyof StopInput, value: string) => {
    const updated = [...stops];
    updated[index] = { ...updated[index], [field]: value };
    setStops(updated);
  };

  const handleAddGuest = () => {
    setGuests([...guests, { name: '', allergies: 'None', dietary_requirements: 'Standard Safari Menu' }]);
  };

  const handleRemoveGuest = (index: number) => {
    if (guests.length <= 1) return;
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

    if (!groupName.trim()) {
      setError('Group name is required.');
      return;
    }

    if (stops.some((s) => !s.lodge_name.trim() || !s.chef_phone.trim() || !s.check_in_date.trim())) {
      setError('All lodge stops must have a lodge name, chef phone number, and valid check-in date.');
      return;
    }

    if (guests.some((g) => !g.name.trim())) {
      setError('All guests must have a valid full name.');
      return;
    }

    setLoading(true);
    try {
      const payload: CreateItineraryRequest = {
        id: customId.trim() ? customId.trim() : undefined,
        group_name: groupName.trim(),
        stops,
        guests,
      };

      const result = await createItinerary(payload);
      onCreated(result);
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to register itinerary';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-stone-900/60 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl border border-stone-200 max-w-4xl w-full max-h-[92vh] flex flex-col overflow-hidden animate-in fade-in duration-200">
        {/* Header */}
        <div className="px-6 py-4 bg-stone-900 text-stone-100 flex items-center justify-between border-b border-stone-800">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 rounded-lg bg-amber-600 flex items-center justify-center text-white">
              <Sparkles className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">Register Safari Itinerary</h2>
              <p className="text-xs text-stone-400">
                Multi-lodge booking with automated 72h privacy-safe SMS triggers &amp; protected allergy manifests
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

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-6 text-sm">
          {error && (
            <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-red-700 flex items-start space-x-2">
              <AlertCircle className="w-5 h-5 shrink-0 mt-0.5 text-red-500" />
              <span>{error}</span>
            </div>
          )}

          {/* Group Details */}
          <div className="bg-stone-50 p-4 rounded-xl border border-stone-200 space-y-4">
            <h3 className="font-semibold text-stone-900 text-base flex items-center space-x-2">
              <span>1. Safari Group Details</span>
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="md:col-span-2">
                <label className="block text-xs font-semibold text-stone-700 mb-1">
                  Safari Group Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={groupName}
                  onChange={(e) => setGroupName(e.target.value)}
                  placeholder="e.g. Serengeti Great Migration Expedition"
                  className="w-full px-3 py-2 rounded-lg border border-stone-300 focus:outline-none focus:ring-2 focus:ring-amber-500 bg-white"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-stone-700 mb-1">
                  Booking Code (Optional)
                </label>
                <input
                  type="text"
                  value={customId}
                  onChange={(e) => setCustomId(e.target.value)}
                  placeholder="e.g. ITIN-2026-901"
                  className="w-full px-3 py-2 rounded-lg border border-stone-300 focus:outline-none focus:ring-2 focus:ring-amber-500 bg-white font-mono text-xs"
                />
              </div>
            </div>
          </div>

          {/* Sequential Lodge Stops */}
          <div className="bg-stone-50 p-4 rounded-xl border border-stone-200 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold text-stone-900 text-base flex items-center space-x-2">
                  <Building2 className="w-4 h-4 text-amber-600" />
                  <span>2. Sequential Lodge Stops</span>
                </h3>
                <p className="text-xs text-stone-500">
                  Chefs will receive a privacy-safe non-sensitive SMS alert exactly 72h prior to arrival.
                </p>
              </div>
              <button
                type="button"
                onClick={handleAddStop}
                className="inline-flex items-center space-x-1 text-xs font-medium bg-amber-100 text-amber-800 px-3 py-1.5 rounded-lg hover:bg-amber-200 transition-colors"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add Lodge Stop</span>
              </button>
            </div>

            <div className="space-y-3">
              {stops.map((stop, index) => (
                <div
                  key={index}
                  className="bg-white p-3.5 rounded-lg border border-stone-200 grid grid-cols-1 md:grid-cols-12 gap-3 items-center"
                >
                  <div className="md:col-span-1 text-xs font-bold text-amber-700 bg-amber-50 w-7 h-7 rounded-full flex items-center justify-center border border-amber-200">
                    #{index + 1}
                  </div>
                  <div className="md:col-span-4">
                    <label className="block text-[11px] font-medium text-stone-600 mb-0.5">Lodge / Camp Name</label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. Mara Serena Lodge"
                      value={stop.lodge_name}
                      onChange={(e) => handleStopChange(index, 'lodge_name', e.target.value)}
                      className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs focus:ring-2 focus:ring-amber-500 focus:outline-none"
                    />
                  </div>
                  <div className="md:col-span-3">
                    <label className="block text-[11px] font-medium text-stone-600 mb-0.5 flex items-center space-x-1">
                      <Phone className="w-3 h-3 text-stone-400" />
                      <span>Chef Phone (GSM)</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="+254711223344"
                      value={stop.chef_phone}
                      onChange={(e) => handleStopChange(index, 'chef_phone', e.target.value)}
                      className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs font-mono focus:ring-2 focus:ring-amber-500 focus:outline-none"
                    />
                  </div>
                  <div className="md:col-span-3">
                    <label className="block text-[11px] font-medium text-stone-600 mb-0.5 flex items-center space-x-1">
                      <Calendar className="w-3 h-3 text-stone-400" />
                      <span>Check-In Date</span>
                    </label>
                    <input
                      type="date"
                      required
                      value={stop.check_in_date}
                      onChange={(e) => handleStopChange(index, 'check_in_date', e.target.value)}
                      className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs focus:ring-2 focus:ring-amber-500 focus:outline-none"
                    />
                  </div>
                  <div className="md:col-span-1 flex justify-end">
                    <button
                      type="button"
                      disabled={stops.length <= 1}
                      onClick={() => handleRemoveStop(index)}
                      className="text-stone-400 hover:text-red-600 disabled:opacity-30 p-1.5 rounded transition-colors"
                      title="Remove lodge stop"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Guest Dietary & Allergy Manifest */}
          <div className="bg-stone-50 p-4 rounded-xl border border-stone-200 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold text-stone-900 text-base flex items-center space-x-2">
                  <User className="w-4 h-4 text-amber-600" />
                  <span>3. Guest Dietary &amp; Allergy Manifest</span>
                </h3>
                <p className="text-xs text-stone-500">
                  Protected health info. Stripped from plain SMS, accessible exclusively via authenticated USSD (*384*44#) or lounge web.
                </p>
              </div>
              <button
                type="button"
                onClick={handleAddGuest}
                className="inline-flex items-center space-x-1 text-xs font-medium bg-amber-100 text-amber-800 px-3 py-1.5 rounded-lg hover:bg-amber-200 transition-colors"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add Guest</span>
              </button>
            </div>

            <div className="space-y-3">
              {guests.map((guest, index) => (
                <div
                  key={index}
                  className="bg-white p-3.5 rounded-lg border border-stone-200 space-y-2.5"
                >
                  <div className="grid grid-cols-1 md:grid-cols-12 gap-3 items-center">
                    <div className="md:col-span-3">
                      <label className="block text-[11px] font-medium text-stone-600 mb-0.5">Guest Full Name</label>
                      <input
                        type="text"
                        required
                        placeholder="e.g. Dr. Jane Goodall"
                        value={guest.name}
                        onChange={(e) => handleGuestChange(index, 'name', e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs focus:ring-2 focus:ring-amber-500 focus:outline-none"
                      />
                    </div>
                    <div className="md:col-span-4">
                      <label className="block text-[11px] font-medium text-stone-600 mb-0.5">
                        Allergies <span className="text-red-500 font-semibold">(Protected Health Data)</span>
                      </label>
                      <input
                        type="text"
                        required
                        placeholder="e.g. Peanut (Severe Anaphylaxis), Shellfish"
                        value={guest.allergies}
                        onChange={(e) => handleGuestChange(index, 'allergies', e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs focus:ring-2 focus:ring-amber-500 focus:outline-none"
                      />
                    </div>
                    <div className="md:col-span-4">
                      <label className="block text-[11px] font-medium text-stone-600 mb-0.5">Dietary Requirements</label>
                      <input
                        type="text"
                        placeholder="e.g. Strict Vegetarian / Halal"
                        value={guest.dietary_requirements || ''}
                        onChange={(e) => handleGuestChange(index, 'dietary_requirements', e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-md border border-stone-300 text-xs focus:ring-2 focus:ring-amber-500 focus:outline-none"
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

                  {/* Quick Select Chips */}
                  <div className="flex flex-wrap items-center gap-1.5 pt-1 text-[11px]">
                    <span className="text-stone-400 text-[10px] uppercase font-semibold">Quick Tags:</span>
                    {COMMON_ALLERGIES.slice(0, 4).map((allergy) => (
                      <button
                        key={allergy}
                        type="button"
                        onClick={() => {
                          const current = guest.allergies === 'None' || !guest.allergies ? '' : guest.allergies;
                          const nextVal = current ? `${current}, ${allergy}` : allergy;
                          handleGuestChange(index, 'allergies', nextVal);
                        }}
                        className="px-2 py-0.5 bg-stone-100 hover:bg-stone-200 text-stone-700 rounded border border-stone-200 text-[10px]"
                      >
                        + {allergy}
                      </button>
                    ))}
                    {COMMON_DIETS.slice(1, 3).map((diet) => (
                      <button
                        key={diet}
                        type="button"
                        onClick={() => handleGuestChange(index, 'dietary_requirements', diet)}
                        className="px-2 py-0.5 bg-amber-50 hover:bg-amber-100 text-amber-800 rounded border border-amber-200 text-[10px]"
                      >
                        {diet}
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
              Total Stops: <span className="font-bold text-stone-800">{stops.length}</span> • Total Manifest Guests: <span className="font-bold text-stone-800">{guests.length}</span>
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
                  <span>Saving Itinerary...</span>
                ) : (
                  <>
                    <Sparkles className="w-4 h-4" />
                    <span>Create Safari Itinerary</span>
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
