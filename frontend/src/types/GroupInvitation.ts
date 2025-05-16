import { Group } from './Group';
import { UserBasicInfo } from './User';

export enum GroupInvitationStatus {
  Pending = 'pending',
  Accepted = 'accepted',
  Rejected = 'rejected',
  Cancelled = 'cancelled', // If inviter cancels
  Expired = 'expired',   // If invitation has a time limit
}

export interface GroupInvitation {
  id: string;
  group_id: string;
  group_name?: string; // Directly available on invitation
  inviter_id: string;
  invitee_id: string;
  status: GroupInvitationStatus;
  created_at: string;
  updated_at: string;

  // Nested details
  group?: Partial<Group>; // Group details (e.g., avatar_url, full name if group_name is just a summary)
  inviter?: UserBasicInfo; // Inviter details
  invitee?: UserBasicInfo; // Invitee details (not used for "invited by" display but part of the structure)
}