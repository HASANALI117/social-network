import React from 'react';
import Tabs from '../../common/Tabs';
import { Text } from '../../ui/text';
import { Button } from '../../ui/button';
// import { Avatar } from '../../ui/avatar'; // No longer used directly
// import GroupPostsTab from './GroupPostsTab'; // Replaced by summary tab
import GroupMembersTab from './GroupMembersTab';
// import GroupEventsTab from './GroupEventsTab'; // Replaced by summary tab
import GroupInviteManager from '../GroupInviteManager';
import GroupJoinRequestsTab from './GroupJoinRequestsTab';
import GroupRecentPostsSummaryTab from './GroupRecentPostsSummaryTab'; // New Summary Tab
import GroupEventsSummaryTab from './GroupEventsSummaryTab'; // New Summary Tab
import { Group } from '../../../types/Group';
import { UserBasicInfo } from '../../../types/User';

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

  const { id: groupId, members_count, posts_count, events_count, members /* posts, events no longer directly used here */ } = group;

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
    {
      label: 'Recent Posts',
      value: 'recent-posts',
      content: <Tabs.Panel id="recent-posts" className="py-4"><GroupRecentPostsSummaryTab groupId={groupId} /></Tabs.Panel>
    },
    {
      label: 'Upcoming Events',
      value: 'upcoming-events',
      content: <Tabs.Panel id="upcoming-events" className="py-4"><GroupEventsSummaryTab groupId={groupId} /></Tabs.Panel>
    },
    {
      label: 'Members',
      value: 'members',
      content: <Tabs.Panel id="members" className="py-4"><GroupMembersTab members={members} /></Tabs.Panel>
    },
    {
      label: 'Invitations',
      value: 'invitations',
      content: invitationsTabContent
    }
  ];

  // Conditionally add the Join Requests tab if the viewer is an admin or creator
  // Assuming viewer_is_admin or similar logic determines if this tab should be shown.
  const isGroupAdmin = group.viewer_is_admin || (currentUser && currentUser.user_id === group.creator_info.user_id);

  if (group?.id && isGroupAdmin) { // Only show to admins/creator
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
          outline
          className="text-sm"
        >
          Leave Group
        </Button>
      </div>

      <Tabs tabs={tabs.map(tab => ({ label: tab.label, value: tab.value, content: tab.content }))} />
      
    </div>
  );
}