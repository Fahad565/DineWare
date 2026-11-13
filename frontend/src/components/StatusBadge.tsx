'use client';

import React from 'react';
import { Clock, Send, CheckCircle2 } from 'lucide-react';
import { StopStatus } from '../types/itinerary';

interface StatusBadgeProps {
  status: StopStatus;
  size?: 'sm' | 'md' | 'lg';
  showIcon?: boolean;
}

export function StatusBadge({ status, size = 'md', showIcon = true }: StatusBadgeProps) {
  const sizeClasses = {
    sm: 'text-xs px-2 py-0.5 gap-1',
    md: 'text-xs sm:text-sm px-2.5 py-1 gap-1.5',
    lg: 'text-sm px-3.5 py-1.5 gap-2 font-semibold',
  };

  if (status === 'confirmed') {
    return (
      <span
        className={`inline-flex items-center rounded-full font-medium border bg-emerald-50 text-emerald-800 border-emerald-300 shadow-sm ${sizeClasses[size]}`}
        title="Chef confirmed manifest receipt via USSD shortcode *384*44# or secure web portal"
      >
        {showIcon && <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600 shrink-0" />}
        <span>Chef Confirmed via USSD</span>
      </span>
    );
  }

  if (status === 'sms_sent') {
    return (
      <span
        className={`inline-flex items-center rounded-full font-medium border bg-sky-50 text-sky-800 border-sky-300 shadow-sm ${sizeClasses[size]}`}
        title="Privacy-safe Africa's Talking SMS alert dispatched 72 hours prior to arrival"
      >
        {showIcon && <Send className="w-3.5 h-3.5 text-sky-600 shrink-0" />}
        <span>SMS Sent (72h Window)</span>
      </span>
    );
  }

  return (
    <span
      className={`inline-flex items-center rounded-full font-medium border bg-amber-50 text-amber-800 border-amber-300 shadow-sm ${sizeClasses[size]}`}
      title="Scheduled for automated dispatch 72 hours before check-in date"
    >
      {showIcon && <Clock className="w-3.5 h-3.5 text-amber-600 shrink-0" />}
      <span>Pending Dispatch</span>
    </span>
  );
}
