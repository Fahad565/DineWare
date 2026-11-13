'use client';

import React, { useState, useEffect, useMemo } from 'react';
import Link from 'next/link';
import {
  Plus,
  Search,
  Filter,
  Users,
  Building2,
  ShieldAlert,
  CheckCircle2,
  Calendar,
  Phone,
  Clock,
  ArrowRight,
  Sparkles,
  ExternalLink,
  Edit3,
  Copy,
  Check,
  Radio,
  RefreshCw,
  UtensilsCrossed,
} from 'lucide-react';
import { Itinerary, getStopStatus, getStatusLabel, StopStatus } from '../types/itinerary';
import { fetchItineraries, checkBackendHealth } from '../lib/api';
import { StatusBadge } from '../components/StatusBadge';
import { CreateItineraryModal } from '../components/CreateItineraryModal';
import { EditGuestsModal } from '../components/EditGuestsModal';

export default function OperatorDashboard() {
  const [itineraries, setItineraries] = useState<Itinerary[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [backendStatus, setBackendStatus] = useState<{ online: boolean; checking: boolean }>({
    online: false,
    checking: true,
  });

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingItinerary, setEditingItinerary] = useState<Itinerary | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [propagationToast, setPropagationToast] = useState<string | null>(null);

  // Load itineraries & test backend connection
  const loadData = React.useCallback(async () => {
    setLoading(true);
    try {
      const [data, health] = await Promise.all([
        fetchItineraries(),
        checkBackendHealth(),
      ]);
      setItineraries(data);
      setBackendStatus({ online: health.online, checking: false });
    } catch (e) {
      console.warn('Failed to fetch data:', e);
      setBackendStatus({ online: false, checking: false });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

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
      severeAllergies += it.guests.filter((g) => {
        const lower = g.allergies.toLowerCase();
        return (
          lower.includes('anaphylaxis') ||
          lower.includes('peanut') ||
          lower.includes('severe') ||
          lower.includes('celiac')
        );
      }).length;
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

  const handleCopyId = (id: string) => {
    navigator.clipboard.writeText(id);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleItineraryCreated = (newItinerary: Itinerary) => {
    setItineraries((prev) => [newItinerary, ...prev]);
    setPropagationToast(`Safari Itinerary "${newItinerary.group_name}" successfully registered! 72h SMS triggers scheduled.`);
    setTimeout(() => setPropagationToast(null), 5000);
  };

  const handleItineraryUpdated = (updated: Itinerary) => {
    setItineraries((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
    setPropagationToast(`Real-time downstream update propagated to all lodge stops for "${updated.group_name}"!`);
    setTimeout(() => setPropagationToast(null), 5000);
  };

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      {/* Top Banner & Title */}
      <div className="bg-gradient-to-r from-stone-900 via-stone-800 to-amber-950 rounded-2xl p-6 sm:p-8 text-white shadow-xl border border-stone-800 relative overflow-hidden">
        <div className="absolute right-0 top-0 bottom-0 w-1/3 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-amber-600/20 via-transparent to-transparent pointer-events-none"></div>

        <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-2 max-w-2xl">
            <div className="flex items-center space-x-2">
              <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-500/20 text-amber-300 border border-amber-500/30">
                Tour Operator Control Center
              </span>
              <div className="flex items-center space-x-1.5 text-xs text-stone-300">
                <Radio className={`w-3.5 h-3.5 ${backendStatus.online ? 'text-emerald-400 animate-pulse' : 'text-amber-400'}`} />
                <span>
                  {backendStatus.online ? 'Backend Engine Connected' : 'Local Fallback Synced'}
                </span>
              </div>
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-white">
              Safari Dietary &amp; Itinerary Synchronization
            </h1>
            <p className="text-stone-300 text-sm leading-relaxed">
              Safeguard safari guests with zero-data GSM signaling. Manage multi-lodge itineraries, track live chef USSD confirmations, and propagate mid-trip allergy updates downstream instantly.
            </p>
          </div>

          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
            <button
              onClick={() => setIsCreateOpen(true)}
              className="px-5 py-3 rounded-xl bg-amber-500 hover:bg-amber-400 text-stone-950 font-bold text-sm shadow-lg hover:shadow-amber-500/20 transition-all flex items-center justify-center space-x-2 group"
            >
              <Plus className="w-4 h-4 transition-transform group-hover:rotate-90" />
              <span>Create Safari Itinerary</span>
            </button>
            <button
              onClick={loadData}
              className="p-3 rounded-xl bg-stone-800 hover:bg-stone-700 text-stone-300 hover:text-white border border-stone-700 transition-colors flex items-center justify-center"
              title="Refresh Itineraries"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {/* Downstream Propagation Toast Banner */}
      {propagationToast && (
        <div className="p-4 rounded-xl bg-emerald-900/90 border border-emerald-600 text-emerald-100 flex items-center justify-between shadow-xl animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="flex items-center space-x-3 text-sm">
            <div className="w-8 h-8 rounded-full bg-emerald-800 flex items-center justify-center shrink-0">
              <Check className="w-4 h-4 text-emerald-300" />
            </div>
            <div>
              <span className="font-bold text-white">Downstream Synchronization Active:</span>{' '}
              <span>{propagationToast}</span>
            </div>
          </div>
          <button
            onClick={() => setPropagationToast(null)}
            className="text-emerald-300 hover:text-white p-1"
          >
            &times;
          </button>
        </div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-5 gap-3 sm:gap-4">
        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm flex items-center space-x-3">
          <div className="w-10 h-10 rounded-lg bg-amber-50 border border-amber-200 flex items-center justify-center text-amber-700 shrink-0">
            <Sparkles className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xl font-bold text-stone-900">{metrics.totalItineraries}</div>
            <div className="text-xs text-stone-500">Active Itineraries</div>
          </div>
        </div>

        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm flex items-center space-x-3">
          <div className="w-10 h-10 rounded-lg bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-700 shrink-0">
            <Building2 className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xl font-bold text-stone-900">{metrics.totalStops}</div>
            <div className="text-xs text-stone-500">Lodge Stops</div>
          </div>
        </div>

        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm flex items-center space-x-3">
          <div className="w-10 h-10 rounded-lg bg-stone-100 border border-stone-200 flex items-center justify-center text-stone-700 shrink-0">
            <Users className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xl font-bold text-stone-900">{metrics.totalGuests}</div>
            <div className="text-xs text-stone-500">Monitored Guests</div>
          </div>
        </div>

        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm flex items-center space-x-3">
          <div className="w-10 h-10 rounded-lg bg-red-50 border border-red-200 flex items-center justify-center text-red-700 shrink-0">
            <ShieldAlert className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xl font-bold text-red-600">{metrics.severeAllergies}</div>
            <div className="text-xs text-stone-500">Severe Allergies</div>
          </div>
        </div>

        <div className="col-span-2 lg:col-span-1 bg-white p-4 rounded-xl border border-stone-200 shadow-sm flex items-center space-x-3">
          <div className="w-10 h-10 rounded-lg bg-emerald-50 border border-emerald-200 flex items-center justify-center text-emerald-700 shrink-0">
            <CheckCircle2 className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xl font-bold text-emerald-700">{metrics.confirmationRate}%</div>
            <div className="text-xs text-stone-500">Chef Confirmed</div>
          </div>
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        <div className="relative flex-1">
          <Search className="w-4 h-4 text-stone-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search by group name, booking code (ITIN-...), lodge, or guest..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-4 py-2 rounded-lg border border-stone-300 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500 bg-stone-50/50"
          />
        </div>

        <div className="flex items-center space-x-2 text-xs">
          <Filter className="w-3.5 h-3.5 text-stone-400 shrink-0" />
          <span className="text-stone-500 font-medium">Filter Status:</span>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 rounded-lg border border-stone-300 bg-white font-medium text-stone-700 focus:outline-none focus:ring-2 focus:ring-amber-500"
          >
            <option value="all">All Stops Status</option>
            <option value="pending">Pending Dispatch</option>
            <option value="sms_sent">SMS Sent (72h Window)</option>
            <option value="confirmed">Chef Confirmed via USSD</option>
          </select>
        </div>
      </div>

      {/* Itinerary Cards */}
      {loading ? (
        <div className="text-center py-16 bg-white rounded-2xl border border-stone-200">
          <RefreshCw className="w-8 h-8 animate-spin mx-auto text-amber-600 mb-3" />
          <p className="text-stone-600 text-sm font-medium">Loading Safari Itineraries...</p>
        </div>
      ) : filteredItineraries.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-2xl border border-dashed border-stone-300 p-8">
          <UtensilsCrossed className="w-12 h-12 mx-auto text-stone-300 mb-3" />
          <h3 className="text-lg font-bold text-stone-800">No Safari Itineraries Found</h3>
          <p className="text-stone-500 text-sm max-w-md mx-auto mt-1 mb-6">
            {searchQuery || statusFilter !== 'all'
              ? 'Try adjusting your search criteria or filter parameters.'
              : 'Register your first multi-lodge safari group to begin automated Africa’s Talking 72h dispatches.'}
          </p>
          <button
            onClick={() => setIsCreateOpen(true)}
            className="inline-flex items-center space-x-2 px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white font-semibold rounded-lg text-sm shadow transition-colors"
          >
            <Plus className="w-4 h-4" />
            <span>Create Itinerary</span>
          </button>
        </div>
      ) : (
        <div className="space-y-6">
          {filteredItineraries.map((itinerary) => (
            <div
              key={itinerary.id}
              className="bg-white rounded-2xl border border-stone-200 shadow-sm hover:shadow-md transition-shadow overflow-hidden"
            >
              {/* Card Header */}
              <div className="p-5 sm:p-6 bg-stone-50 border-b border-stone-200 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="space-y-1">
                  <div className="flex items-center space-x-2">
                    <span className="font-mono text-xs font-bold bg-stone-200 text-stone-800 px-2.5 py-0.5 rounded-md flex items-center space-x-1">
                      <span>{itinerary.id}</span>
                      <button
                        onClick={() => handleCopyId(itinerary.id)}
                        className="hover:text-amber-700 ml-1"
                        title="Copy Booking Reference"
                      >
                        {copiedId === itinerary.id ? (
                          <Check className="w-3 h-3 text-emerald-600" />
                        ) : (
                          <Copy className="w-3 h-3 text-stone-500" />
                        )}
                      </button>
                    </span>
                    <span className="text-xs text-stone-500">
                      Created: {new Date(itinerary.created_at).toLocaleDateString()}
                    </span>
                    {itinerary.updated_at && (
                      <span className="text-xs text-amber-700 bg-amber-50 px-2 py-0.5 rounded border border-amber-200 font-medium">
                        Updated Mid-Trip
                      </span>
                    )}
                  </div>
                  <h2 className="text-lg sm:text-xl font-bold text-stone-900 tracking-tight">
                    {itinerary.group_name}
                  </h2>
                </div>

                <div className="flex items-center space-x-2 sm:space-x-3">
                  <button
                    onClick={() => setEditingItinerary(itinerary)}
                    className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-white border border-stone-300 hover:border-amber-500 text-stone-700 hover:text-amber-800 text-xs font-semibold shadow-sm transition-all"
                    title="Edit guest counts or allergies mid-trip and demonstrate real-time downstream propagation"
                  >
                    <Edit3 className="w-3.5 h-3.5 text-amber-600" />
                    <span>Edit Guests / Allergies</span>
                  </button>
                  <Link
                    href={`/manifest/${itinerary.id}`}
                    className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-stone-900 hover:bg-stone-800 text-amber-300 text-xs font-semibold shadow-sm transition-colors"
                  >
                    <span>Lodge Manifest View</span>
                    <ExternalLink className="w-3.5 h-3.5" />
                  </Link>
                </div>
              </div>

              {/* Sequential Lodge Stops Visualizer */}
              <div className="p-5 sm:p-6 border-b border-stone-100">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-xs font-bold text-stone-400 uppercase tracking-wider flex items-center space-x-1.5">
                    <Building2 className="w-3.5 h-3.5 text-stone-500" />
                    <span>Sequential Safari Lodge Stops ({itinerary.stops.length} properties)</span>
                  </h3>
                  <span className="text-xs text-stone-500">
                    Dispatched 72h ahead via Africa’s Talking SMS
                  </span>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {itinerary.stops.map((stop, idx) => {
                    const status = getStopStatus(stop);
                    return (
                      <div
                        key={stop.id || idx}
                        className="p-4 rounded-xl border border-stone-200 bg-stone-50/50 hover:bg-white hover:border-stone-300 transition-all space-y-3 relative flex flex-col justify-between"
                      >
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="text-[11px] font-bold text-amber-800 bg-amber-100/60 px-2 py-0.5 rounded-full">
                              Stop #{idx + 1}
                            </span>
                            <StatusBadge status={status} size="sm" />
                          </div>

                          <h4 className="font-bold text-stone-900 text-sm leading-snug">
                            {stop.lodge_name}
                          </h4>

                          <div className="space-y-1 text-xs text-stone-600 pt-1">
                            <div className="flex items-center space-x-2">
                              <Calendar className="w-3.5 h-3.5 text-stone-400 shrink-0" />
                              <span className="font-medium text-stone-800">{stop.check_in_date}</span>
                            </div>
                            <div className="flex items-center space-x-2">
                              <Phone className="w-3.5 h-3.5 text-stone-400 shrink-0" />
                              <span className="font-mono text-stone-700">{stop.chef_phone}</span>
                            </div>
                          </div>
                        </div>

                        <div className="pt-2 border-t border-stone-200/60 flex items-center justify-between text-[11px]">
                          <span className="text-stone-400">
                            {status === 'confirmed'
                              ? 'Confirmed via *384*44#'
                              : status === 'sms_sent'
                              ? 'Alert sent to chef'
                              : 'Scheduled 72h window'}
                          </span>
                          <Link
                            href={`/manifest/${itinerary.id}?lodge=${encodeURIComponent(stop.lodge_name)}`}
                            className="text-amber-700 hover:text-amber-900 font-semibold flex items-center space-x-1"
                          >
                            <span>Lounge View</span>
                            <ArrowRight className="w-3 h-3" />
                          </Link>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Guest Manifest Summary & Allergy Tags */}
              <div className="p-5 sm:p-6 bg-white flex flex-col md:flex-row items-start md:items-center justify-between gap-4 text-xs">
                <div className="space-y-2 max-w-3xl">
                  <div className="flex items-center space-x-2">
                    <Users className="w-3.5 h-3.5 text-stone-400" />
                    <span className="font-bold text-stone-700">
                      Guest Roster ({itinerary.guests.length} Travelers):
                    </span>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    {itinerary.guests.map((guest, gIdx) => {
                      const hasSevere =
                        guest.allergies.toLowerCase().includes('anaphylaxis') ||
                        guest.allergies.toLowerCase().includes('peanut') ||
                        guest.allergies.toLowerCase().includes('severe');
                      const hasAllergy = guest.allergies && guest.allergies !== 'None';

                      return (
                        <div
                          key={guest.id || gIdx}
                          className={`px-2.5 py-1 rounded-lg border text-xs flex items-center space-x-1.5 ${
                            hasSevere
                              ? 'bg-red-50 border-red-200 text-red-900 font-medium'
                              : hasAllergy
                              ? 'bg-amber-50 border-amber-200 text-amber-900'
                              : 'bg-stone-50 border-stone-200 text-stone-700'
                          }`}
                        >
                          {hasSevere && <ShieldAlert className="w-3 h-3 text-red-600 shrink-0" />}
                          <span className="font-semibold">{guest.name}:</span>
                          <span className="text-[11px] opacity-90">{guest.allergies}</span>
                          {guest.dietary_requirements && guest.dietary_requirements !== 'Standard' && (
                            <span className="text-[10px] bg-white/70 px-1 rounded font-mono text-stone-600">
                              {guest.dietary_requirements}
                            </span>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>

                <div className="text-right shrink-0">
                  <button
                    onClick={() => setEditingItinerary(itinerary)}
                    className="text-xs text-amber-700 hover:text-amber-900 font-semibold underline underline-offset-2 flex items-center space-x-1"
                  >
                    <span>Demonstrate Downstream Propagation</span>
                    <ArrowRight className="w-3 h-3" />
                  </button>
                </div>
              </div>
            </div>
          ))}
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

      {/* Zero-Data Architecture Explainer Footnote */}
      <div className="bg-stone-100 rounded-2xl p-6 border border-stone-200 text-xs text-stone-600 space-y-3">
        <h4 className="font-bold text-stone-800 text-sm flex items-center space-x-2">
          <span>Zero-Data Safari Synchronization Architecture</span>
        </h4>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-white p-3.5 rounded-lg border border-stone-200 space-y-1">
            <span className="font-bold text-stone-900 block text-xs">1. Staggered 72h SMS Alerts</span>
            <p>
              The Go backend monitors departure dates. 72 hours prior to arrival, Africa’s Talking SMS alerts the head chef with itinerary ID and headcount. No cleartext allergy data is ever sent via SMS.
            </p>
          </div>
          <div className="bg-white p-3.5 rounded-lg border border-stone-200 space-y-1">
            <span className="font-bold text-stone-900 block text-xs">2. Offline USSD Lookup (*384*44#)</span>
            <p>
              Camp chefs in dead cellular zones dial standard GSM shortcode <code className="bg-stone-100 px-1 py-0.5 rounded text-amber-700 font-mono">*384*44#</code>. It pulls live allergy specs from Neon PostgreSQL with zero internet data.
            </p>
          </div>
          <div className="bg-white p-3.5 rounded-lg border border-stone-200 space-y-1">
            <span className="font-bold text-stone-900 block text-xs">3. Secure Lounge Manifest View</span>
            <p>
              Properties with guest lounge Wi-Fi access authenticated online manifests at <code className="bg-stone-100 px-1 py-0.5 rounded text-amber-700 font-mono">/manifest/[code]</code> to review specs and click &quot;Confirm Manifest&quot;.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
