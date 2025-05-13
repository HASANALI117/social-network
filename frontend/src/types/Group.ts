// frontend/src/types/Group.ts
export interface Group {
  id: string;
  creator_id: string;
  title: string;
  description: string;
  created_at: string; // ISO date string
  updated_at: string; // ISO date string
  // Optional fields that might be populated by the backend later (e.g., member count, creator details)
  creator_first_name?: string;
  creator_last_name?: string;
  member_count?: number;
}