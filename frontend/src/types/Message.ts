// frontend/src/types/Message.ts
export interface Message {
  id?: string; // Optional: Client-side ID before backend confirmation or if backend provides it
  type: 'direct' | 'group';
  sender_id: string;
  receiver_id: string; // Target user ID for direct messages
  content: string;
  created_at: string; // ISO 8601 format (e.g., from new Date().toISOString())
  sender_username?: string; // Optional: For display purposes
  sender_avatar_url?: string; // Optional: For display purposes
}