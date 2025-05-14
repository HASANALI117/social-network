import React, { useState, useEffect, useCallback } from 'react';
import Tabs from '../../common/Tabs';
import { Text, Strong } from '../../ui/text';
import { Button } from '../../ui/button';
import { Avatar } from '../../ui/avatar';
import GroupPostsTab from './GroupPostsTab';
import GroupMembersTab from './GroupMembersTab';
import GroupEventsTab from './GroupEventsTab';
import GroupInviteManager from '../GroupInviteManager';
import GroupJoinRequestsTab from './GroupJoinRequestsTab'; // New import
import { Group } from '../../../types/Group';
import { UserBasicInfo } from '../../../types/User';
// GroupJoinRequest, useRequest, toast are no longer needed here as they are handled by GroupJoinRequestsTab

interface GroupMemberViewProps {
  group: Group;
  currentUser: UserBasicInfo | null;
  handleLeaveGroup: () => void;
}

export default function GroupMemberView({
  group,
  currentUser,
  handleLeaveGroup,
}: GroupMemberViewProps) {

  const { id: groupId, members_count, posts_count, events_count, posts, members, events, creator_info, viewer_is_admin } = group;

  const handleInviteSent = (invitedUserId: string, successMessage: string) => {
    // Optionally, update UI or state here, e.g., refetch members
  };

  const handleInviteError = (errorMessage: string) => {
    // Error already toasted by GroupInviteManager
  };

  const invitationsTabContent = (
    <Tabs.Panel id="invitations" className="py-4 space-y-6">
      <div>
          <GroupInviteManager
            groupId={groupId}
            currentUser={currentUser}
            onInviteSent={handleInviteSent}
            onInviteError={handleInviteError}
          />
      </div>
    </Tabs.Panel>
  );

  const tabs = [
    { label: 'Posts', value: 'posts', content: <Tabs.Panel id="posts" className="py-4"><GroupPostsTab posts={posts} /></Tabs.Panel> },
    { label: 'Members', value: 'members', content: <Tabs.Panel id="members" className="py-4"><GroupMembersTab members={members} /></Tabs.Panel> },
    { label: 'Events', value: 'events', content: <Tabs.Panel id="events" className="py-4"><GroupEventsTab events={events} /></Tabs.Panel> },
    { label: 'Invitations', value: 'invitations', content: invitationsTabContent }
  ];

  // Unconditionally add the Join Requests tab
  // Ensure group.id is available and valid here
  if (group?.id) {
    tabs.push({
        label: 'Join Requests',
        value: 'join-requests',
        content: <Tabs.Panel id="join-requests" className="py-4"><GroupJoinRequestsTab groupId={group.id} /></Tabs.Panel>,
    });
  }

  return (
    <div>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8 text-lg text-center">
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{members_count}</Text>
          <Text className="text-gray-300">Members</Text>
        </div>
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{posts_count}</Text>
          <Text className="text-gray-300">Posts</Text>
        </div>
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{events_count}</Text>
          <Text className="text-gray-300">Events</Text>
        </div>
      </div>
      
      <div className="mb-6 flex justify-end">
         <Button
          onClick={handleLeaveGroup}
          outline // Secondary action style
          className="text-sm" // Adjust size as needed
        >
          Leave Group
        </Button>
      </div>

      <Tabs tabs={tabs.map(tab => ({ label: tab.label, content: tab.content }))} />
      
    </div>
  );
}