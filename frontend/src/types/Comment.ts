// frontend/src/types/Comment.ts
export interface Comment {
  id: string;
  post_id: string;
  user_id: string;
  content: string;
  image_url?: string;
  created_at: string; // ISO date string
  user_first_name?: string;
  user_last_name?: string;
  user_avatar_url?: string;
  username?: string;
}