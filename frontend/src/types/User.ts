import { Post } from './Post';

// Base user type matching backend UserResponse
export interface User {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  about_me?: string;
  birth_date: string;
  is_private: boolean;
  created_at: string;
  updated_at: string;
}

// Summary for user, used in FollowRequest
export interface UserSummary {
  id: string;
  username: string;
  avatar_url?: string;
  first_name: string;
  last_name: string;
}

// Basic user info for lists like group members
export interface UserBasicInfo {
  user_id: string; // Changed from id to user_id to match group.creator_info and typical API responses for nested user data
  username: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

export interface FollowRequest {
  id: string;
  requester: UserSummary;
  target: UserSummary;
  status: 'pending' | 'accepted' | 'declined';
  created_at: string;
  // Note: The components for managing follow requests (ManageFollowRequestsSection,
  // FollowRequestList, FollowRequestCard) when consuming data from the
  // `/api/users/me/follow-requests` endpoint now operate directly on `UserSummary[]`
  // (for `received` and `sent` lists) rather than this `FollowRequest` structure.
  // This `FollowRequest` type might be used for other API responses or contexts
  // where a full request object with requester, target, and status is provided.
}

// Extended user profile type matching backend UserProfileResponse
export interface UserProfile extends User {
  followers_count: number;
  following_count: number;
  latest_posts: Post[];
  latest_followers: User[];
  latest_following: User[];
  is_followed: boolean;
  follow_request_state?: 'SENT' | 'RECEIVED' | '';
  // is_private is already in User interface
}

// Type for user signup data
export interface UserSignupData {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
  birth_date: string;
  username?: string;
  about_me?: string;
  avatar_url?: string | null;
}

// Type for updating user profile data
export interface UpdateUserProfileData {
  first_name?: string;
  last_name?: string;
  username?: string;
  about_me?: string;
  avatar_url?: string;
  // Add other updatable fields as needed, making them optional
}
