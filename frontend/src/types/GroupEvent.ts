import { UserBasicInfo } from './User';

export interface GroupEventResponseOption {
  id: string;
  text: string;
  count: number;
  users?: UserBasicInfo[]; // List of users who chose this option
}

export interface GroupEvent {
  id: string;
  group_id: string;
  creator_id: string;
  creator_info: UserBasicInfo;
  title: string;
  description: string;
  event_time: string; // ISO 8601 date-time
  created_at: string;
  updated_at: string;
  options: GroupEventResponseOption[];
  current_user_response_id?: string; // ID of the option user selected
}