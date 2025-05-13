// frontend/src/types/GroupJoinRequest.ts
export type GroupJoinRequestStatus = "pending" | "accepted" | "declined";

export interface GroupJoinRequest {
  id: string;
  group_id: string;
  user_id: string;
  status: GroupJoinRequestStatus;
  created_at: string; // ISO date string
  updated_at: string; // ISO date string
  // Optional populated fields for display
  group_title?: string;
  user_name?: string;
  user_avatar_url?: string;
}