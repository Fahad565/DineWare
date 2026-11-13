export interface Stop {
  id?: number;
  itinerary_id?: string;
  lodge_name: string;
  chef_phone: string;
  check_in_date: string;
  notification_sent?: boolean;
  chef_confirmed?: boolean;
}

export interface Guest {
  id?: number;
  itinerary_id?: string;
  name: string;
  allergies: string;
  dietary_requirements?: string;
}

export interface Itinerary {
  id: string;
  group_name: string;
  created_at: string;
  updated_at?: string;
  stops: Stop[];
  guests: Guest[];
}

export interface StopInput {
  lodge_name: string;
  chef_phone: string;
  check_in_date: string;
}

export interface GuestInput {
  name: string;
  allergies: string;
  dietary_requirements?: string;
}

export interface CreateItineraryRequest {
  id?: string;
  group_name: string;
  stops: StopInput[];
  guests: GuestInput[];
}

export interface UpdateGuestsRequest {
  guests: GuestInput[];
}

export type StopStatus = 'pending' | 'sms_sent' | 'confirmed';

export function getStopStatus(stop: Stop): StopStatus {
  if (stop.chef_confirmed) {
    return 'confirmed';
  }
  if (stop.notification_sent) {
    return 'sms_sent';
  }
  return 'pending';
}

export function getStatusLabel(status: StopStatus): string {
  switch (status) {
    case 'confirmed':
      return 'Chef Confirmed via USSD';
    case 'sms_sent':
      return 'SMS Sent (72h Window)';
    case 'pending':
    default:
      return 'Pending Dispatch';
  }
}
