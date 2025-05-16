import { UserBasicInfo } from './User';

// Assuming 'responses' is an array of individual user responses
// We need to define what an individual response looks like.
// For now, let's assume it contains user info and their chosen option text/id.
export interface IndividualEventResponse {
  user_id: string;
  username: string;
  first_name?: string;
  last_name?: string;
  avatar_url?: string;
  response: string; // e.g., "going", "not_going"
  updated_at: string;
}

export interface GroupEvent { // Original GroupEvent for list views
  id: string;
  group_id: string;
  creator_id: string;
  title: string;
  description: string;
  event_time: string; // ISO 8601 date-time
  created_at: string;
  updated_at: string;
  creator_name?: string;
  group_name?: string;
  response_count?: number; // Optional: might be useful for list views
}

export interface GroupEventDetail { // New detailed view type
  id: string;
  group_id: string;
  creator_id: string;
  title: string;
  description: string;
  event_time: string; // ISO 8601 date-time
  created_at: string;
  updated_at: string;
  creator_name?: string;
  creator_info?: UserBasicInfo;
  group_name?: string;
  group_avatar_url?: string;
  responses: IndividualEventResponse[];
  response_counts: { [key: string]: number }; // e.g., { "going": 0, "not_going": 0 }
  current_user_response_option_id?: string;
}

// Original EventOption and EventResponse for reference, might be deprecated or reused differently
export interface EventOption {
  id: string;
  event_id: string;
  option_text: string;
  created_at: string;
  updated_at: string;
}

export interface EventResponse {
  id: string;
  event_id: string;
  user_id: string;
  option_id: string;
  created_at: string;
  updated_at: string;
  user_info?: UserBasicInfo; // Added for convenience
  option_text?: string; // Added for convenience
}

// This structure was based on the old API and needs to be updated
// to reflect the new `responses` and `response_counts`
export interface GroupEventWithDetails extends GroupEvent {
  options: Array<EventOption & { count: number; users: UserBasicInfo[] }>;
  current_user_response_option_id?: string;
}