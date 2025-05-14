import React from 'react';
import Link from 'next/link';
import { Avatar } from '../../ui/avatar';
import { Text } from '../../ui/text';
import { UserBasicInfo } from '../../../types/User';

interface GroupMembersTabProps {
  members: UserBasicInfo[] | undefined;
}

export default function GroupMembersTab({ members }: GroupMembersTabProps) {
  if (!members || members.length === 0) {
    return <Text className="text-center text-gray-400 py-4">No members to display.</Text>;
  }

  return (
    <div className="space-y-3">
      {members.map((member: UserBasicInfo) => (
        <div key={member.user_id} className="flex items-center p-3 bg-gray-700 rounded-lg shadow">
          <Avatar
            src={member.avatar_url || null}
            initials={!member.avatar_url && member.first_name && member.last_name ? `${member.first_name.substring(0, 1)}${member.last_name.substring(0, 1)}`.toUpperCase() : undefined}
            alt={`${member.first_name} ${member.last_name}`}
            className="h-10 w-10 mr-3 rounded-full"
          />
          <div>
            <Link href={`/profile/${member.user_id}`} className="text-purple-300 hover:underline font-semibold">
              {member.first_name} {member.last_name}
            </Link>
            <Text className="text-xs text-gray-400">@{member.username}</Text>
          </div>
        </div>
      ))}
    </div>
  );
}