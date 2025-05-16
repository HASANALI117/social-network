// frontend/src/types/Group.ts
import { User, UserBasicInfo } from './User'; // UserBasicInfo will be added to User.ts
import { Post } from './Post';

export interface PostSummary {
  id: string;
  content: string;
  image_url?: string;
  created_at: string;
  creator_id: string;
  creator_name: string; // e.g., "John Doe"
  creator_avatar_url?: string;
}

export interface EventSummary {
  id: string;
  title: string;
  description: string;
  event_time?: string; // date-time string
  // Add other relevant summary fields if needed
}

export interface Group { // This represents GroupDetailResponse
  id: string;
  name: string;
  description: string;
  avatar_url?: string;
  creator_info: {
    user_id: string;
    username: string;
    first_name: string;
    last_name: string;
    avatar_url?: string;
  };
  created_at: string; // date-time string
  members_count: number;
  posts_count: number;
  events_count: number;
  viewer_pending_request_id?: string;
  viewer_pending_request_status?: 'pending' | 'accepted' | 'rejected' | null;
viewer_is_admin?: boolean;
  // Member-specific fields (conditionally present)
  members?: UserBasicInfo[];
  posts?: PostSummary[];
  events?: EventSummary[];
}