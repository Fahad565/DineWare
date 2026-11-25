'use client';

import React, { useState, useEffect, useMemo } from 'react';
import Link from 'next/link';
import { useParams, useSearchParams } from 'next/navigation';
import {
  ShieldAlert,
  CheckCircle2,
  Calendar,
  Phone,
  Building2,
  Users,
  Utensils,
  ArrowLeft,
  Printer,
  Wifi,
  Sparkles,
  AlertTriangle,
  RefreshCw,
  FileCheck,
  PhoneCall,
  Check,
} from 'lucide-react';
import { Itinerary, Stop, Guest, getStopStatus } from '../../../types/itinerary';
import { fetchItinerary } from '../../../lib/api';
import { StatusBadge } from '../../../components/StatusBadge';

export default function LodgeManifestPage() {
  const params = useParams();
  const searchParams = useSearchParams();

  const code = (params?.code as string) || '';
  const initialLodgeQuery = searchParams.get('lodge') || '';

  const [itinerary, setItinerary] = useState<Itinerary | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [selectedLodgeName, setSelectedLodgeName] = useState<string>(initialLodgeQuery);

  // Load manifest
  const loadManifest = React.useCallback(async () => {
    if (!code) return;
    setLoading(true);
    setLoadError(null);
    try {
      const data = await fetchItinerary(code);
      setItinerary(data);
      if (data && data.stops.length > 0 && !selectedLodgeName) {
        // default to first lodge or query match
        const match = initialLodgeQuery
          ? data.stops.find((s) => s.lodge_name.toLowerCase() === initialLodgeQuery.toLowerCase())
          : data.stops[0];
        setSelectedLodgeName(match ? match.lodge_name : data.stops[0].lodge_name);
      }
    } catch (e) {
      setItinerary(null);
      setLoadError(e instanceof Error ? e.message : 'Failed to load manifest from the backend.');
    } finally {
      setLoading(false);
    }
  }, [code, initialLodgeQuery, selectedLodgeName]);

  useEffect(() => {
    loadManifest();
  }, [loadManifest]);

  // Active lodge stop
  const currentStop: Stop | undefined = useMemo(() => {
    if (!itinerary || itinerary.stops.length === 0) return undefined;
    if (selectedLodgeName) {
      const found = itinerary.stops.find(
        (s) => s.lodge_name.toLowerCase() === selectedLodgeName.toLowerCase()
      );
      if (found) return found;
    }
    return itinerary.stops[0];
  }, [itinerary, selectedLodgeName]);

  // Aggregate stats
  const manifestStats = useMemo(() => {
    if (!itinerary) return { total: 0, severe: 0, dietary: 0, standard: 0 };
    const total = itinerary.guests.length;
    let severe = 0;
    let dietary = 0;
    let standard = 0;

    itinerary.guests.forEach((g) => {
      const lower = g.allergies.toLowerCase();
      const isSevere =
        lower.includes('anaphylaxis') ||
        lower.includes('peanut') ||
        lower.includes('severe') ||
        lower.includes('celiac');
      if (isSevere) {
        severe++;
      }
      if (
        (g.dietary_requirements &&
          g.dietary_requirements.toLowerCase() !== 'standard' &&
          g.dietary_requirements.toLowerCase() !== 'standard safari menu') ||
        (g.allergies && g.allergies !== 'None' && !isSevere)
      ) {
        dietary++;
      }
      if (
        (!g.allergies || g.allergies === 'None') &&
        (!g.dietary_requirements ||
          g.dietary_requirements.toLowerCase() === 'standard' ||
          g.dietary_requirements.toLowerCase() === 'standard safari menu')
      ) {
        standard++;
      }
    });

    return { total, severe, dietary, standard };
  }, [itinerary]);

  const handlePrint = () => {
    if (typeof window !== 'undefined') {
      window.print();
    }
  };

  const isStopConfirmed = Boolean(currentStop?.chef_confirmed);

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto py-16 text-center">
        <RefreshCw className="w-10 h-10 animate-spin text-amber-600 mx-auto mb-4" />
        <h2 className="text-xl font-bold text-stone-800">Retrieving Secure Kitchen Manifest...</h2>
        <p className="text-stone-500 text-sm mt-1">Decrypting guest health and dietary profiles for {code}</p>
      </div>
    );
  }

  if (!itinerary) {
    return (
      <div className="max-w-xl mx-auto py-16 text-center space-y-4">
        <div className="w-16 h-16 rounded-full bg-red-100 text-red-600 flex items-center justify-center mx-auto">
          <ShieldAlert className="w-8 h-8" />
        </div>
        <h2 className="text-2xl font-bold text-stone-900">
          {loadError ? 'Could Not Load Manifest' : 'Manifest Not Found'}
        </h2>
        <p className="text-stone-600 text-sm">
          {loadError
            ? loadError
            : <>No safari itinerary was found for booking reference code <code className="bg-stone-200 px-1.5 py-0.5 rounded font-mono text-stone-900 font-bold">{code}</code>.</>}
        </p>
        {loadError && (
          <button
            onClick={loadManifest}
            className="inline-flex items-center gap-2 rounded-lg bg-amber-700 px-4 py-2.5 text-sm font-semibold text-white hover:bg-amber-800"
          >
            <RefreshCw className="h-4 w-4" />
            Retry
          </button>
        )}
        <div className="pt-4">
          <Link
            href="/"
            className="inline-flex items-center space-x-2 px-5 py-2.5 rounded-xl bg-stone-900 hover:bg-stone-800 text-white text-sm font-semibold transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Return to Tour Operator Dashboard</span>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto space-y-8 animate-in fade-in duration-300">
      {/* Top Nav Breadcrumb & Terminal Status */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <Link
          href="/"
          className="inline-flex items-center space-x-2 text-stone-600 hover:text-stone-900 font-semibold text-sm transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to Tour Operator Dashboard</span>
        </Link>

        <div className="flex items-center space-x-2">
          <div className="flex items-center space-x-1.5 px-3 py-1 rounded-full bg-emerald-50 text-emerald-800 border border-emerald-300 text-xs font-semibold">
            <Wifi className="w-3.5 h-3.5 text-emerald-600 animate-pulse" />
            <span>Lounge Wi-Fi Terminal Online</span>
          </div>
          <button
            onClick={handlePrint}
            className="inline-flex items-center space-x-1.5 px-3 py-1 rounded-lg bg-white border border-stone-300 hover:border-stone-400 text-stone-700 text-xs font-semibold shadow-sm transition-colors"
          >
            <Printer className="w-3.5 h-3.5 text-stone-600" />
            <span>Print Kitchen Prep Sheet</span>
          </button>
        </div>
      </div>

      {/* Manifest Header Banner */}
      <div className="bg-stone-900 text-stone-100 rounded-2xl p-6 sm:p-8 shadow-xl border border-stone-800 relative overflow-hidden">
        <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-2">
            <div className="flex flex-wrap items-center gap-2">
              <span className="font-mono text-xs font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30 px-2.5 py-0.5 rounded-md">
                Ref: {itinerary.id}
              </span>
              {currentStop && (
                <StatusBadge status={getStopStatus(currentStop)} size="sm" />
              )}
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              {currentStop?.lodge_name || 'Safari Lodge Kitchen'}
            </h1>
            <p className="text-stone-300 text-sm">
              Safari Group: <span className="text-white font-semibold">{itinerary.group_name}</span>
            </p>
          </div>

          {/* Lodge Selector if multi-lodge */}
          {itinerary.stops.length > 1 && (
            <div className="bg-stone-800/90 p-3 rounded-xl border border-stone-700 text-xs space-y-1.5 shrink-0">
              <label className="block text-stone-400 font-semibold uppercase text-[10px] tracking-wider">
                Select Lodge Property Stop:
              </label>
              <select
                value={selectedLodgeName}
                onChange={(e) => setSelectedLodgeName(e.target.value)}
                className="w-full bg-stone-900 border border-stone-700 text-stone-100 px-3 py-1.5 rounded-lg text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-amber-500"
              >
                {itinerary.stops.map((stop, idx) => (
                  <option key={stop.id || idx} value={stop.lodge_name}>
                    Stop #{idx + 1}: {stop.lodge_name} ({stop.check_in_date})
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>

        {/* Current Stop Logistics Bar */}
        {currentStop && (
          <div className="mt-6 pt-5 border-t border-stone-800 grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs text-stone-300">
            <div className="flex items-center space-x-2">
              <Calendar className="w-4 h-4 text-amber-400 shrink-0" />
              <div>
                <span className="text-stone-400 block text-[10px] uppercase">Scheduled Arrival</span>
                <span className="font-semibold text-white text-sm">{currentStop.check_in_date}</span>
              </div>
            </div>
            <div className="flex items-center space-x-2">
              <Phone className="w-4 h-4 text-amber-400 shrink-0" />
              <div>
                <span className="text-stone-400 block text-[10px] uppercase">Kitchen Head Chef GSM</span>
                <span className="font-mono font-semibold text-white text-sm">{currentStop.chef_phone}</span>
              </div>
            </div>
            <div className="flex items-center space-x-2">
              <Users className="w-4 h-4 text-amber-400 shrink-0" />
              <div>
                <span className="text-stone-400 block text-[10px] uppercase">Total Group Size</span>
                <span className="font-semibold text-white text-sm">{itinerary.guests.length} Arriving Guests</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Confirmation Banner if confirmed */}
      {isStopConfirmed && (
        <div className="p-4 rounded-xl bg-emerald-50 border border-emerald-300 text-emerald-900 flex items-start sm:items-center justify-between gap-4 shadow-sm animate-in fade-in duration-300">
          <div className="flex items-center space-x-3">
            <div className="w-9 h-9 rounded-full bg-emerald-600 text-white flex items-center justify-center shrink-0">
              <Check className="w-5 h-5" />
            </div>
            <div className="space-y-0.5">
              <div className="font-bold text-sm text-emerald-950 flex items-center space-x-2">
                <span>Manifest Receipt Confirmed</span>
                <span className="text-[10px] bg-emerald-200/80 text-emerald-900 px-2 py-0.5 rounded font-mono font-semibold">
                  AUDITED
                </span>
              </div>
              <p className="text-xs text-emerald-800">
                Kitchen confirmed receipt of dietary &amp; allergy manifests. Synchronized with Tour Operator Dashboard.
              </p>
            </div>
          </div>
          <span className="text-xs font-semibold text-emerald-800 uppercase tracking-wider shrink-0 hidden sm:block">
            Status: Confirmed
          </span>
        </div>
      )}

      {/* Critical Allergy & Anaphylaxis Warning Banner */}
      {manifestStats.severe > 0 && (
        <div className="bg-red-50 border-2 border-red-300 rounded-2xl p-5 shadow-sm space-y-3">
          <div className="flex items-start space-x-3">
            <div className="w-10 h-10 rounded-xl bg-red-600 text-white flex items-center justify-center shrink-0 shadow">
              <ShieldAlert className="w-6 h-6" />
            </div>
            <div className="space-y-1">
              <div className="flex items-center space-x-2">
                <span className="text-xs uppercase font-extrabold tracking-wider bg-red-600 text-white px-2 py-0.5 rounded">
                  Critical Health Alert
                </span>
                <span className="text-sm font-bold text-red-950">
                  {manifestStats.severe} Guest(s) with Life-Threatening Severe Allergies
                </span>
              </div>
              <p className="text-xs text-red-800 leading-relaxed">
                Back-of-house kitchen teams must sanitize dedicated prep tables, utilize designated separate cookware, and ensure cross-contamination controls are strictly enforced before meal preparation.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Kitchen Meal Prep Breakdown Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm space-y-1">
          <span className="text-xs font-bold text-stone-500 uppercase tracking-wider">Total Headcount</span>
          <div className="text-2xl font-black text-stone-900">{manifestStats.total}</div>
          <p className="text-[11px] text-stone-400">Total dining covers</p>
        </div>

        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm space-y-1">
          <span className="text-xs font-bold text-red-600 uppercase tracking-wider">Severe Allergies</span>
          <div className="text-2xl font-black text-red-600">{manifestStats.severe}</div>
          <p className="text-[11px] text-red-400">Anaphylaxis / strict risk</p>
        </div>

        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm space-y-1">
          <span className="text-xs font-bold text-amber-700 uppercase tracking-wider">Dietary Needs</span>
          <div className="text-2xl font-black text-amber-700">{manifestStats.dietary}</div>
          <p className="text-[11px] text-amber-500">Vegetarian, Halal, etc.</p>
        </div>

        <div className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm space-y-1">
          <span className="text-xs font-bold text-stone-600 uppercase tracking-wider">Standard Meals</span>
          <div className="text-2xl font-black text-stone-800">{manifestStats.standard}</div>
          <p className="text-[11px] text-stone-400">Standard safari menu</p>
        </div>
      </div>

      {/* Detailed Guest Manifest Table */}
      <div className="bg-white rounded-2xl border border-stone-200 shadow-sm overflow-hidden">
        <div className="px-6 py-4 bg-stone-50 border-b border-stone-200 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Utensils className="w-4 h-4 text-amber-600" />
            <h2 className="text-base font-bold text-stone-900">
              Individual Guest Dietary &amp; Allergy Specifications
            </h2>
          </div>
          <span className="text-xs text-stone-500">
            Protected Medical Data • Authenticated Terminal
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-stone-700">
            <thead className="bg-stone-100/75 text-stone-600 text-xs uppercase font-semibold border-b border-stone-200">
              <tr>
                <th className="py-3 px-4">#</th>
                <th className="py-3 px-4">Guest Name</th>
                <th className="py-3 px-4">Allergy Manifest</th>
                <th className="py-3 px-4">Dietary Requirements</th>
                <th className="py-3 px-4">Kitchen Action &amp; Safety Protocol</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-stone-200">
              {itinerary.guests.map((guest, idx) => {
                const lower = guest.allergies.toLowerCase();
                const isSevere =
                  lower.includes('anaphylaxis') ||
                  lower.includes('peanut') ||
                  lower.includes('severe') ||
                  lower.includes('celiac');
                const hasAllergy = guest.allergies && guest.allergies !== 'None';

                return (
                  <tr
                    key={guest.id || idx}
                    className={`hover:bg-stone-50/75 transition-colors ${
                      isSevere ? 'bg-red-50/30' : ''
                    }`}
                  >
                    <td className="py-3.5 px-4 font-mono text-xs text-stone-400">{idx + 1}</td>
                    <td className="py-3.5 px-4 font-bold text-stone-900">
                      {guest.name}
                    </td>
                    <td className="py-3.5 px-4">
                      {isSevere ? (
                        <div className="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-md bg-red-100 text-red-900 font-bold text-xs border border-red-300">
                          <ShieldAlert className="w-3.5 h-3.5 text-red-600 shrink-0" />
                          <span>{guest.allergies}</span>
                        </div>
                      ) : hasAllergy ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-md bg-amber-100 text-amber-900 font-medium text-xs border border-amber-200">
                          {guest.allergies}
                        </span>
                      ) : (
                        <span className="text-stone-400 text-xs">No known allergies</span>
                      )}
                    </td>
                    <td className="py-3.5 px-4 font-medium text-stone-800">
                      {guest.dietary_requirements || 'Standard Safari Menu'}
                    </td>
                    <td className="py-3.5 px-4 text-xs">
                      {isSevere ? (
                        <span className="font-bold text-red-700">
                          Strict sterile surface. Do not use shared frying oil or cutting boards.
                        </span>
                      ) : hasAllergy ? (
                        <span className="text-amber-800">
                          Check all ingredient labels and pre-packed condiments.
                        </span>
                      ) : (
                        <span className="text-stone-500">Standard preparation routine.</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Confirmation Action Section */}
      <div className="bg-stone-50 rounded-2xl p-6 sm:p-8 border border-stone-200 flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
        <div className="space-y-1.5 max-w-xl">
          <h3 className="text-base font-bold text-stone-900 flex items-center space-x-2">
            <FileCheck className="w-5 h-5 text-amber-600" />
            <span>Kitchen Manifest Acknowledgment Loop</span>
          </h3>
          <p className="text-xs text-stone-600 leading-relaxed">
            Chef confirmations are recorded by the backend after the chef confirms receipt in the secure USSD workflow.
          </p>
        </div>

        <div className="flex items-center space-x-3 shrink-0">
          {isStopConfirmed ? (
            <div className="inline-flex items-center space-x-2 px-5 py-3 rounded-xl bg-emerald-600 text-white font-bold text-sm shadow-md">
              <CheckCircle2 className="w-5 h-5" />
              <span>Manifest Confirmed</span>
            </div>
          ) : (
            <span className="text-sm font-semibold text-amber-800">Chef confirmation pending</span>
          )}
        </div>
      </div>

      {/* Zero-Data Offline USSD Guidance */}
      <div className="bg-white rounded-2xl p-5 border border-stone-200 flex items-start space-x-4 shadow-sm text-xs text-stone-600">
        <div className="w-10 h-10 rounded-xl bg-stone-100 flex items-center justify-center shrink-0 text-amber-700">
          <PhoneCall className="w-5 h-5" />
        </div>
        <div className="space-y-1">
          <h4 className="font-bold text-stone-900 text-sm">
            Offline GSM Kitchen Fallback: Dial *384*95853#
          </h4>
          <p>
            If satellite lounge Wi-Fi goes down or cellular data is unavailable, the head chef can dial{' '}
            <code className="bg-stone-100 text-amber-800 font-mono font-bold px-1.5 py-0.5 rounded border border-stone-300">
              *384*95853#
            </code>{' '}
            from phone <code className="font-mono text-stone-900">{currentStop?.chef_phone || '+2547...'}</code>. Enter booking code{' '}
            <code className="bg-stone-100 text-amber-800 font-mono font-bold px-1.5 py-0.5 rounded border border-stone-300">
              {itinerary.id}
            </code>{' '}
            to query allergy specs and confirm receipt over zero-data GSM signaling channels.
          </p>
        </div>
      </div>
    </div>
  );
}
