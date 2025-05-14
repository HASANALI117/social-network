import React from 'react';
import { Heading } from '../../ui/heading';
import { Text } from '../../ui/text';
import { Button } from '../../ui/button';

interface GroupNonMemberViewProps {
  members_count: number;
  posts_count: number;
  events_count: number;
  handleRequestToJoin: () => void;
}

export default function GroupNonMemberView({
  members_count,
  posts_count,
  events_count,
  handleRequestToJoin,
}: GroupNonMemberViewProps) {
  return (
    <div className="text-center">
      <Heading level={2} className="text-2xl mb-4 text-gray-200">
        Join the Conversation!
      </Heading>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6 text-lg">
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
      <Button
        onClick={handleRequestToJoin}
        className="bg-purple-600 hover:bg-purple-700 text-white font-bold py-3 px-6 rounded-lg text-lg"
      >
        Request to Join
      </Button>
    </div>
  );
}