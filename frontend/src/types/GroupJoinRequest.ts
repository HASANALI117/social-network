import { UserBasicInfo } from './User';

export type GroupJoinRequestStatus = "pending" | "accepted" | "rejected";

export interface GroupJoinRequest {
  id: string;
  group_id: string;
  requester_id: string;
  status: GroupJoinRequestStatus;
  created_at: string;
  updated_at: string;
  requester?: UserBasicInfo; 
}