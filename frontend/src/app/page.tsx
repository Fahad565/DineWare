'use client';

import React, { useState, useEffect, useMemo, useRef } from 'react';
import Link from 'next/link';
import {
  Plus,
  Search,
  Filter,
  Users,
  Building2,
  Calendar,
  Phone,
  ArrowRight,
  ExternalLink,
  Edit3,
  Pencil,
  Trash2,
  Copy,
  Check,
  RefreshCw,
  UtensilsCrossed,
  AlertTriangle,
  ShieldAlert,
} from 'lucide-react';
import { Itinerary, getStopStatus } from '../types/itinerary';
import { deleteItinerary, fetchItineraries, checkBackendHealth } from '../lib/api';
import { StatusBadge } from '../components/StatusBadge';
import { CreateItineraryModal } from '../components/CreateItineraryModal';
import { EditItineraryModal } from '../components/EditItineraryModal';
import { EditGuestsModal } from '../components/EditGuestsModal';

function formatCheckInDate(value: string) {
  const date = new Date(`${value}T12:00:00`);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat('en', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }).format(date);
}

function hasSevereAllergy(allergies: string) {
  const lower = allergies.toLowerCase();
  return (
    lower.includes('anaphylaxis') ||
    lower.includes('peanut') ||
    lower.includes('severe') ||
    lower.includes('celiac')
  );
}

function countMealNotes(itinerary: Itinerary) {
  return itinerary.guests.filter((guest) => (
    (guest.allergies && guest.allergies !== 'None') ||
    (guest.dietary_requirements &&
      !['standard', 'standard safari menu'].includes(guest.dietary_requirements.toLowerCase()))
  )).length;
}

function countSevereAllergies(itinerary: Itinerary) {
  return itinerary.guests.filter((guest) => hasSevereAllergy(guest.allergies || '')).length;
}

export default function OperatorDashboard() {
  const [itineraries, setItineraries] = useState<Itinerary[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [backendStatus, setBackendStatus] = useState<{ online: boolean; checking: boolean }>({
    online: false,
    checking: true,
  });

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingItinerary, setEditingItinerary] = useState<Itinerary | null>(null);
  const [editingTrip, setEditingTrip] = useState<Itinerary | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [propagationToast, setPropagationToast] = useState<string | null>(null);
  const toastTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Load itineraries & test backend connection
  const loadData = React.useCallback(async () => {
    setLoading(true);
    try {
      const [data, health] = await Promise.all([
        fetchItineraries(),
        checkBackendHealth(),
      ]);
      setItineraries(data);
      setLoadError(null);
      setBackendStatus({ online: health.online, checking: false });
    } catch (e) {
      setItineraries([]);
      setLoadError(e instanceof Error ? e.message : 'Failed to load itineraries from the backend.');
      setBackendStatus({ online: false, checking: false });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  useEffect(() => {
    return () => {
      if (toastTimeoutRef.current) clearTimeout(toastTimeoutRef.current);
      if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current);
    };
  }, []);

  // Filtered list
  const filteredItineraries = useMemo(() => {
    return itineraries.filter((item) => {
      const matchesSearch =
        item.group_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.stops.some((s) => s.lodge_name.toLowerCase().includes(searchQuery.toLowerCase())) ||
        item.guests.some((g) => g.name.toLowerCase().includes(searchQuery.toLowerCase()));

      if (!matchesSearch) return false;

      if (statusFilter === 'all') return true;

      // Filter by stop status
      return item.stops.some((s) => getStopStatus(s) === statusFilter);
    });
  }, [itineraries, searchQuery, statusFilter]);

  // Global aggregate metrics
  const metrics = useMemo(() => {
    const totalItineraries = itineraries.length;
    let totalStops = 0;
    let confirmedStops = 0;
    let totalGuests = 0;
    let severeAllergies = 0;

    itineraries.forEach((it) => {
      totalStops += it.stops.length;
      confirmedStops += it.stops.filter((s) => s.chef_confirmed).length;
      totalGuests += it.guests.length;
      severeAllergies += countSevereAllergies(it);
    });

    const confirmationRate = totalStops > 0 ? Math.round((confirmedStops / totalStops) * 100) : 0;

    return {
      totalItineraries,
      totalStops,
      totalGuests,
      severeAllergies,
      confirmationRate,
    };
  }, [itineraries]);

  const showPropagationToast = (message: string) => {
    if (toastTimeoutRef.current) clearTimeout(toastTimeoutRef.current);
    setPropagationToast(message);
    toastTimeoutRef.current = setTimeout(() => setPropagationToast(null), 5000);
  };

  const handleCopyId = (id: string) => {
    navigator.clipboard.writeText(id);
    setCopiedId(id);
    if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current);
    copyTimeoutRef.current = setTimeout(() => setCopiedId(null), 2000);
  };

  const handleItineraryCreated = (newItinerary: Itinerary) => {
    setItineraries((prev) => [newItinerary, ...prev]);
    showPropagationToast(`“${newItinerary.group_name}” has been added.`);
  };

  const handleItineraryUpdated = (updated: Itinerary) => {
    setItineraries((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
    showPropagationToast(`“${updated.group_name}” has been updated.`);
  };

  const handleDeleteItinerary = async (itinerary: Itinerary) => {
    if (!window.confirm('Delete this itinerary and all lodge stops?')) return;

    setActionError(null);
    setDeletingId(itinerary.id);
    try {
      await deleteItinerary(itinerary.id);
      setItineraries((current) => current.filter((item) => item.id !== itinerary.id));
      setEditingItinerary((current) => current?.id === itinerary.id ? null : current);
      setEditingTrip((current) => current?.id === itinerary.id ? null : current);
      showPropagationToast(`“${itinerary.group_name}” and its lodge stops have been deleted.`);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to delete itinerary.');
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      <section className="flex flex-col gap-5 rounded-2xl border border-stone-200 bg-white p-5 shadow-sm sm:flex-row sm:items-center sm:justify-between sm:p-7">
        <div className="max-w-2xl">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium text-stone-500">
            <span className="rounded-full bg-stone-100 px-2.5 py-1 text-stone-700">Safari operations</span>
            <span className="flex items-center gap-1.5">
              <span className={`h-2 w-2 rounded-full ${backendStatus.online ? 'bg-emerald-600' : backendStatus.checking ? 'bg-amber-500' : 'bg-stone-400'}`} />
              {backendStatus.checking ? 'Checking connection' : backendStatus.online ? 'Connected' : 'Using saved trips'}
            </span>
          </div>
          <h1 className="font-display text-2xl font-bold tracking-tight text-stone-900">
            Your safari itineraries
          </h1>
          <p className="mt-1.5 text-sm leading-6 text-stone-600">
            Keep lodge teams up to date on guest arrivals and meal requirements.
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <button
            onClick={() => setIsCreateOpen(true)}
            className="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-amber-700 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-amber-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber-700 focus-visible:ring-offset-2"
          >
            <Plus className="h-4 w-4" />
            <span>New itinerary</span>
          </button>
          <button
            onClick={loadData}
            className="inline-flex h-11 w-11 items-center justify-center rounded-lg border border-stone-300 bg-white text-stone-600 transition-colors hover:bg-stone-50 hover:text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber-700 focus-visible:ring-offset-2"
            title="Refresh itineraries"
            aria-label="Refresh itineraries"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </section>

      {propagationToast && (
        <div role="status" className="flex items-center justify-between rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          <div className="flex items-center gap-2">
            <Check className="h-4 w-4 shrink-0 text-emerald-700" />
            <span>{propagationToast}</span>
          </div>
          <button
            onClick={() => {
              if (toastTimeoutRef.current) clearTimeout(toastTimeoutRef.current);
              setPropagationToast(null);
            }}
            className="rounded p-1 text-emerald-800 hover:bg-emerald-100"
            aria-label="Dismiss notification"
          >
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
      )}

      {actionError && (
        <div role="alert" className="flex items-center justify-between gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          <span>{actionError}</span>
          <button
            onClick={() => setActionError(null)}
            className="rounded px-2 py-1 font-medium hover:bg-red-100"
          >
            Dismiss
          </button>
        </div>
      )}

      <div className="grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-stone-200 bg-stone-100 sm:grid-cols-3 lg:grid-cols-5">
        {[
          { label: 'Active trips', value: metrics.totalItineraries },
          { label: 'Lodge stops', value: metrics.totalStops },
          { label: 'Guests', value: metrics.totalGuests },
          { label: 'Allergy alerts', value: metrics.severeAllergies },
          { label: 'Chef confirmations', value: `${metrics.confirmationRate}%` },
        ].map((metric, index) => (
          <div
            key={metric.label}
            className={`min-w-0 bg-white px-4 py-4 sm:px-5 ${index === 4 ? 'col-span-2 sm:col-span-1' : ''}`}
          >
            <p className="text-2xl font-semibold tracking-tight text-stone-900">{metric.value}</p>
            <p className="mt-1 truncate text-xs text-stone-500">{metric.label}</p>
          </div>
        ))}
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-col gap-3 rounded-xl border border-stone-200 bg-white p-3 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-stone-400" />
          <input
            type="text"
            aria-label="Search itineraries"
            placeholder="Search trips, lodges or guests"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full rounded-lg border border-transparent bg-stone-50 py-2.5 pl-9 pr-4 text-sm text-stone-900 placeholder:text-stone-400 focus:border-amber-700 focus:bg-white focus:outline-none focus:ring-2 focus:ring-amber-700/15"
          />
        </div>

        <div className="flex items-center gap-2 sm:pl-3">
          <Filter className="h-4 w-4 shrink-0 text-stone-400" />
          <select
            aria-label="Filter by stop status"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="min-h-10 flex-1 rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-700 focus:border-amber-700 focus:outline-none focus:ring-2 focus:ring-amber-700/15 sm:flex-none"
          >
            <option value="all">All statuses</option>
            <option value="pending">Scheduled</option>
            <option value="sms_sent">SMS sent</option>
            <option value="confirmed">Confirmed</option>
          </select>
        </div>
      </div>

      {/* Itinerary Cards */}
      {loading ? (
        <div className="text-center py-16 bg-white rounded-2xl border border-stone-200">
          <RefreshCw className="mx-auto mb-3 h-7 w-7 animate-spin text-amber-700" />
          <p className="text-sm font-medium text-stone-600">Loading itineraries…</p>
        </div>
      ) : loadError ? (
        <div role="alert" className="rounded-2xl border border-red-200 bg-red-50 px-6 py-10 text-center">
          <h3 className="font-display text-lg font-bold text-red-900">Could not load live itineraries</h3>
          <p className="mx-auto mt-2 max-w-xl text-sm text-red-800">{loadError}</p>
          <button
            onClick={loadData}
            className="mt-5 inline-flex items-center gap-2 rounded-lg bg-red-800 px-4 py-2.5 text-sm font-semibold text-white hover:bg-red-900"
          >
            <RefreshCw className="h-4 w-4" />
            Retry
          </button>
        </div>
      ) : filteredItineraries.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-stone-300 bg-white px-6 py-14 text-center">
          <UtensilsCrossed className="mx-auto mb-3 h-10 w-10 text-stone-300" />
          <h3 className="font-display text-lg font-bold text-stone-800">
            {searchQuery || statusFilter !== 'all' ? 'No matching trips' : 'No itineraries yet'}
          </h3>
          <p className="mx-auto mb-6 mt-1 max-w-md text-sm text-stone-500">
            {searchQuery || statusFilter !== 'all'
              ? 'Try another search or choose a different status.'
              : 'Add a trip to keep lodge teams informed about arrivals and meal requirements.'}
          </p>
          <button
            onClick={() => setIsCreateOpen(true)}
            className="inline-flex items-center gap-2 rounded-lg bg-amber-700 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-amber-800"
          >
            <Plus className="h-4 w-4" />
            <span>New itinerary</span>
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredItineraries.map((itinerary) => {
            const mealNoteCount = countMealNotes(itinerary);
            const severeAllergyCount = countSevereAllergies(itinerary);

            return (
              <div
                key={itinerary.id}
                className="overflow-hidden rounded-2xl border border-stone-200 bg-white"
              >
                <div className="flex flex-col gap-4 border-b border-stone-100 px-5 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-6">
                  <div className="min-w-0">
                    <div className="mb-1.5 flex flex-wrap items-center gap-2">
                      <span className="inline-flex items-center gap-1.5 rounded-md bg-stone-100 px-2 py-1 font-mono text-[11px] font-medium text-stone-600">
                        <span>{itinerary.id}</span>
                        <button
                          onClick={() => handleCopyId(itinerary.id)}
                          className="rounded p-0.5 text-stone-500 hover:bg-stone-200 hover:text-stone-900"
                          title="Copy booking reference"
                          aria-label={`Copy booking reference ${itinerary.id}`}
                        >
                          {copiedId === itinerary.id ? (
                            <Check className="h-3 w-3 text-emerald-700" />
                          ) : (
                            <Copy className="h-3 w-3" />
                          )}
                        </button>
                      </span>
                      {itinerary.updated_at && (
                        <span className="rounded-full bg-amber-50 px-2 py-1 text-[11px] font-medium text-amber-800">
                          Updated
                        </span>
                      )}
                      {severeAllergyCount > 0 && (
                        <span className="inline-flex items-center gap-1 rounded-full bg-red-50 px-2 py-1 text-[11px] font-medium text-red-800">
                          <ShieldAlert className="h-3 w-3 shrink-0" />
                          {severeAllergyCount} severe {severeAllergyCount === 1 ? 'allergy' : 'allergies'}
                        </span>
                      )}
                    </div>
                    <h2 className="truncate font-display text-lg font-bold tracking-tight text-stone-900 sm:text-xl">
                      {itinerary.group_name}
                    </h2>
                    <p className="mt-1 text-xs text-stone-500">
                      {itinerary.stops.length} {itinerary.stops.length === 1 ? 'lodge' : 'lodges'} · {itinerary.guests.length} {itinerary.guests.length === 1 ? 'guest' : 'guests'}
                    </p>
                  </div>

                  <div className="flex shrink-0 flex-wrap items-center gap-2">
                    <button
                      onClick={() => setEditingTrip(itinerary)}
                      className="inline-flex min-h-9 items-center gap-1.5 rounded-lg border border-stone-300 bg-white px-3 py-2 text-xs font-medium text-stone-700 transition-colors hover:border-stone-400 hover:bg-stone-50"
                      aria-label={`Edit trip ${itinerary.group_name}`}
                    >
                      <Pencil className="h-3.5 w-3.5" />
                      <span>Edit trip</span>
                    </button>
                    <button
                      onClick={() => setEditingItinerary(itinerary)}
                      className="inline-flex min-h-9 items-center gap-1.5 rounded-lg border border-stone-300 bg-white px-3 py-2 text-xs font-medium text-stone-700 transition-colors hover:border-stone-400 hover:bg-stone-50"
                    >
                      <Edit3 className="h-3.5 w-3.5" />
                      <span>Edit guests</span>
                    </button>
                    <button
                      onClick={() => handleDeleteItinerary(itinerary)}
                      disabled={deletingId === itinerary.id}
                      className="inline-flex min-h-9 items-center gap-1.5 rounded-lg border border-red-200 bg-white px-3 py-2 text-xs font-medium text-red-700 transition-colors hover:border-red-300 hover:bg-red-50 disabled:cursor-wait disabled:opacity-60"
                      aria-label={`Delete trip ${itinerary.group_name}`}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      <span>{deletingId === itinerary.id ? 'Deleting…' : 'Delete'}</span>
                    </button>
                    <Link
                      href={`/manifest/${itinerary.id}`}
                      className="inline-flex min-h-9 items-center gap-1.5 rounded-lg bg-stone-900 px-3 py-2 text-xs font-medium text-white transition-colors hover:bg-stone-700"
                    >
                      <span>Open manifest</span>
                      <ExternalLink className="h-3.5 w-3.5" />
                    </Link>
                  </div>
                </div>

                <div className="px-5 py-4 sm:px-6">
                  <div className="mb-3 flex items-center justify-between gap-3">
                    <h3 className="flex items-center gap-2 text-sm font-medium text-stone-800">
                      <Building2 className="h-4 w-4 text-stone-500" />
                      Lodge route
                    </h3>
                    <span className="hidden text-xs text-stone-500 sm:block">
                      Kitchen teams receive updates before arrival
                    </span>
                  </div>

                  <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
                    {itinerary.stops.map((stop, idx) => {
                      const status = getStopStatus(stop);
                      return (
                        <div
                          key={stop.id || idx}
                          className="flex min-w-0 flex-col rounded-xl border border-stone-200 p-4"
                        >
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-mono text-xs font-medium text-stone-500">
                              Stop {idx + 1}
                            </span>
                            <StatusBadge status={status} size="sm" />
                          </div>

                          <h4 className="mt-3 line-clamp-2 min-h-10 font-display text-base font-bold leading-5 text-stone-900">
                            {stop.lodge_name}
                          </h4>

                          <div className="mt-3 space-y-1.5 text-xs text-stone-600">
                            <div className="flex items-center gap-2">
                              <Calendar className="h-3.5 w-3.5 shrink-0 text-stone-400" />
                              <span>Arrives {formatCheckInDate(stop.check_in_date)}</span>
                            </div>
                            <div className="flex items-center gap-2">
                              <Phone className="h-3.5 w-3.5 shrink-0 text-stone-400" />
                              <span className="truncate font-mono text-stone-700">{stop.chef_phone}</span>
                            </div>
                          </div>

                          <div className="mt-4 flex items-center justify-between gap-2 border-t border-stone-100 pt-3">
                            <span className="flex min-w-0 items-center gap-1.5 text-xs text-stone-500">
                              {mealNoteCount > 0 && <AlertTriangle className="h-3.5 w-3.5 shrink-0 text-amber-700" />}
                              <span className="truncate">
                                {mealNoteCount > 0
                                  ? `${mealNoteCount} guests with meal notes`
                                  : 'No meal notes'}
                              </span>
                            </span>
                            <Link
                              href={`/manifest/${itinerary.id}?lodge=${encodeURIComponent(stop.lodge_name)}`}
                              className="inline-flex shrink-0 items-center gap-1 text-xs font-medium text-amber-800 hover:text-amber-950"
                            >
                              <span>Details</span>
                              <ArrowRight className="h-3 w-3" />
                            </Link>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>

                <div className="flex flex-col gap-2 border-t border-stone-100 bg-stone-50/70 px-5 py-3.5 text-xs sm:flex-row sm:items-center sm:justify-between sm:px-6">
                  <div className="flex items-center gap-2 text-stone-600">
                    <Users className="h-3.5 w-3.5 shrink-0 text-stone-400" />
                    <span>{itinerary.guests.length} {itinerary.guests.length === 1 ? 'guest' : 'guests'}</span>
                    <span aria-hidden="true">·</span>
                    <span>Meal details are available in the manifest</span>
                  </div>
                  <span className="text-stone-500">Updates are sent before each arrival</span>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Modals */}
      <CreateItineraryModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onCreated={handleItineraryCreated}
      />

      <EditGuestsModal
        isOpen={!!editingItinerary}
        onClose={() => setEditingItinerary(null)}
        itinerary={editingItinerary}
        onUpdated={handleItineraryUpdated}
      />

      <EditItineraryModal
        isOpen={!!editingTrip}
        onClose={() => setEditingTrip(null)}
        itinerary={editingTrip}
        onUpdated={handleItineraryUpdated}
      />

    </div>
  );
}
