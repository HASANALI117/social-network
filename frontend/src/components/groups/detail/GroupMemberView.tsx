import React from 'react';
import Tabs from '../../common/Tabs';
import { Text } from '../../ui/text';
import { Button } from '../../ui/button';
import GroupPostsTab from './GroupPostsTab';
import GroupMembersTab from './GroupMembersTab';
import GroupEventsTab from './GroupEventsTab';
import GroupInviteUserUI from './GroupInviteUserUI';
import { Group, PostSummary, EventSummary } from '../../../types/Group';
import { UserBasicInfo } from '../../../types/User';

interface GroupMemberViewProps {
  group: Group; // Full group object for stats and tab content
  showInviteUI: boolean;
  toggleInviteUI: () => void;
  handleLeaveGroup: () => void;
  // Props for GroupInviteUserUI
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  isActualSearchLoading: boolean;
  searchError: string | null;
  setSearchError: (error: string | null) => void;
  inviteSuccess: string | null;
  setInviteSuccess: (message: string | null) => void;
  inviteError: string | null;
  setInviteError: (error: string | null) => void;
  searchResults: UserBasicInfo[];
  handleSendInvite: (userIdToInvite: string) => void;
  isInviteHookLoading: boolean;
  invitingUserId: string | null;
}

export default function GroupMemberView({
  group,
  showInviteUI,
  toggleInviteUI,
  handleLeaveGroup,
  searchTerm,
  setSearchTerm,
  isActualSearchLoading,
  searchError,
  setSearchError,
  inviteSuccess,
  setInviteSuccess,
  inviteError,
  setInviteError,
  searchResults,
  handleSendInvite,
  isInviteHookLoading,
  invitingUserId,
}: GroupMemberViewProps) {
  const { members_count, posts_count, events_count, posts, members, events } = group;

  return (
    <div>
      {/* Stats remain visible for members */}
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

      <Tabs tabs={[
        {
          label: 'Posts',
          content: (
            <Tabs.Panel id="posts" className="py-4">
              <GroupPostsTab posts={posts} />
            </Tabs.Panel>
          )
        },
        {
          label: 'Members',
          content: (
            <Tabs.Panel id="members" className="py-4">
              <GroupMembersTab members={members} />
            </Tabs.Panel>
          )
        },
        {
          label: 'Events',
          content: (
            <Tabs.Panel id="events" className="py-4">
              <GroupEventsTab events={events} />
            </Tabs.Panel>
          )
        },
        {
          label: 'Invitations',
          content: (
            <Text className="text-center text-gray-400 py-4">
              Group Invitations and Join Requests management will go here.
            </Text>
          )
        }
      ]} />

      <div className="flex flex-col sm:flex-row justify-center space-y-3 sm:space-y-0 sm:space-x-4 mt-8">
        <Button
          onClick={toggleInviteUI}
          className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-md"
        >
          {showInviteUI ? 'Close Invite UI' : 'Invite User'}
        </Button>
        <Button
          onClick={handleLeaveGroup}
          color="red"
          className="bg-red-600 hover:bg-red-700 text-white font-semibold py-2 px-4 rounded-md"
        >
          Leave Group
        </Button>
      </div>

      {showInviteUI && (
        <GroupInviteUserUI
          searchTerm={searchTerm}
          setSearchTerm={setSearchTerm}
          isActualSearchLoading={isActualSearchLoading}
          searchError={searchError}
          setSearchError={setSearchError}
          inviteSuccess={inviteSuccess}
          setInviteSuccess={setInviteSuccess}
          inviteError={inviteError}
          setInviteError={setInviteError}
          searchResults={searchResults}
          handleSendInvite={handleSendInvite}
          isInviteHookLoading={isInviteHookLoading}
          invitingUserId={invitingUserId}
        />
      )}
    </div>
  );
}