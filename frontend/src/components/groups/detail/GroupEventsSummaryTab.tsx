import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRequest } from '../../../hooks/useRequest';
import { GroupEvent } from '../../../types/GroupEvent';
import { GroupEventCard } from '../events/GroupEventCard';
import { Button } from '../../ui/button';
import { Text } from '../../ui/text';

interface GroupEventsSummaryTabProps {
  groupId: string;
}

interface GroupEventResponse {
  events: GroupEvent[];
}

const GroupEventsSummaryTab: React.FC<GroupEventsSummaryTabProps> = ({ groupId }) => {
  const { data, isLoading, error, get } = useRequest<GroupEventResponse>();
  const [events, setEvents] = useState<GroupEvent[]>([]);
  
  useEffect(() => {
    get(`/api/groups/${groupId}/events?limit=3&sort=upcoming`);
  }, [get, groupId]);

  useEffect(() => {
    if (data) {
      setEvents(data.events);
    }
  }, [data]);

  if (isLoading) {
    return <Text>Loading upcoming events...</Text>;
  }

  if (error) {
    return <Text color="red">Error loading events: {error.message}</Text>;
  }

  if (!events || !Array.isArray(events) || events.length === 0) {
    return (
      <div className="text-center py-4">
        <Text>No upcoming events to display.</Text>
        <div className="mt-4">
          <Button href={`/groups/${groupId}/events`} outline>
            View All Events
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {events.map((event) => (
        <GroupEventCard key={event.id} event={event} groupId={groupId} onEventUpdated={() => { /* Placeholder for summary view */ }} />
      ))}
      <div className="mt-6 text-center">
        <Button href={`/groups/${groupId}/events`} color="purple">
          View All Events
        </Button>
      </div>
    </div>
  );
};

export default GroupEventsSummaryTab;