// frontend/src/types/GroupInvitation.ts
export type GroupInvitationStatus = "pending" | "accepted" | "declined";

export interface GroupInvitation {
  id: string;
  group_id: string;
  inviter_id: string;
  invitee_id: string;
  status: GroupInvitationStatus;
  created_at: string; // ISO date string
  updated_at: string; // ISO date string
  // Optional populated fields for display
  group_title?: string;
  inviter_name?: string;
  invitee_name?: string;
}