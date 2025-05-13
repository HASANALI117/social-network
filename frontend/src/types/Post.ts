export interface Post {
  id: string;
  user_id: string;
  group_id?: string | null; // Match backend (sql.NullString becomes string | null)
  title: string;
  content: string;
  image_url?: string; // Optional
  privacy: 'public' | 'semi_private' | 'private'; // Match backend model constants
  created_at: string; // ISO date string from backend
  user_first_name?: string; // Optional, from PostResponse
  user_last_name?: string;  // Optional, from PostResponse
  user_avatar_url?: string; // Optional, from PostResponse
}