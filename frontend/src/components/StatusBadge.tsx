'use client';

import React from 'react';
import { Clock3, Send, Check } from 'lucide-react';
import { StopStatus } from '../types/itinerary';

interface StatusBadgeProps {
  status: StopStatus;
  size?: 'sm' | 'md' | 'lg';
  showIcon?: boolean;
}

export function StatusBadge({ status, size = 'md', showIcon = true }: StatusBadgeProps) {
  const sizeClasses = {
    sm: 'text-[11px] px-2 py-1 gap-1',
    md: 'text-xs px-2.5 py-1 gap-1.5',
    lg: 'text-sm px-3 py-1.5 gap-2 font-semibold',
  };

  if (status === 'confirmed') {
    return (
      <span
        className={`inline-flex items-center rounded-full font-sans font-medium border bg-emerald-50 text-emerald-800 border-emerald-200 ${sizeClasses[size]}`}
        title="Chef confirmed manifest receipt via USSD shortcode *384*44# or secure web portal"
      >
        {showIcon && <Check className="w-3.5 h-3.5 text-emerald-700 shrink-0" />}
        <span>Confirmed</span>
      </span>
    );
  }

  if (status === 'sms_sent') {
    return (
      <span
        className={`inline-flex items-center rounded-full font-sans font-medium border bg-sky-50 text-sky-800 border-sky-200 ${sizeClasses[size]}`}
        title="Privacy-safe Africa's Talking SMS alert dispatched 72 hours prior to arrival"
      >
        {showIcon && <Send className="w-3.5 h-3.5 text-sky-600 shrink-0" />}
        <span>SMS sent</span>
      </span>
    );
  }

  return (
    <span
      className={`inline-flex items-center rounded-full font-sans font-medium border bg-amber-50 text-amber-800 border-amber-200 ${sizeClasses[size]}`}
      title="Scheduled for automated dispatch 72 hours before check-in date"
    >
      {showIcon && <Clock3 className="w-3.5 h-3.5 text-amber-700 shrink-0" />}
      <span>Scheduled</span>
    </span>
  );
}
