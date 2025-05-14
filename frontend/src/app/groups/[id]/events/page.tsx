'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { useRequest } from '@/hooks/useRequest';
import { GroupEvent } from '@/types/GroupEvent';
import { Group } from '@/types/Group'; // For group details
import { CreateGroupEventForm } from '@/components/groups/events/CreateGroupEventForm';
import { GroupEventList } from '@/components/groups/events/GroupEventList';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { Button } from '@/components/ui/button'; // For potential refresh or other actions

interface GroupEventsApiResponse {
  events: GroupEvent[];
  limit: number;
  offset: number;
  count: number;
}

export default function GroupEventsPage() {
  const params = useParams();
  const groupId = params.id as string;

  const [events, setEvents] = useState<GroupEvent[]>([]);
  const [group, setGroup] = useState<Group | null>(null);

  // Fetch group details
  const { get: getGroup, error: groupError, isLoading: groupLoading } = useRequest<Group>();
  // Fetch events
  const {
    get: getEvents,
    error: eventsError,
    isLoading: eventsLoading,
  } = useRequest<GroupEventsApiResponse>();

  const fetchGroupDetails = useCallback(async () => {
    if (groupId) {
      const fetchedGroup = await getGroup(`/api/groups/${groupId}`);
      if (fetchedGroup) {
        setGroup(fetchedGroup);
      }
    }
  }, [groupId, getGroup]);

  const fetchEvents = useCallback(async () => {
    if (groupId) {
      const fetchedEvents = await getEvents(`/api/groups/${groupId}/events`);
      // Ensure fetchedEvents is an array before setting state,
      // or set to an empty array if it's not an array (e.g., null, undefined, or empty object from API)
      setEvents(
        Array.isArray(fetchedEvents!.events) ? fetchedEvents!.events : []
      );
    }
  }, [groupId, getEvents]);

  useEffect(() => {
    fetchGroupDetails();
    fetchEvents();
  }, [fetchGroupDetails, fetchEvents]);

  const handleEventCreated = (newEvent: GroupEvent) => {
    // Ensure newEvent.options is an array, defaulting to an empty array if not.
    const eventWithOptions = {
      ...newEvent,
      options: Array.isArray(newEvent.options) ? newEvent.options : [],
    };
    setEvents((prevEvents) => [eventWithOptions, ...prevEvents]);
  };

  const handleEventUpdated = (updatedEvent: GroupEvent) => {
    setEvents((prevEvents) =>
      prevEvents.map((event) => (event.id === updatedEvent.id ? updatedEvent : event))
    );
  };

  if (groupLoading) return <Text>Loading group details...</Text>;
  if (groupError) return <Text className="text-red-500">Error loading group: {groupError.message}</Text>;

  return (
    <div className="container mx-auto p-4">
      <Heading level={1} className="mb-6">
        Events for {group ? `"${group.name}"` : 'Group'}
      </Heading>

      {groupId && (
        <div className="mb-6">
          <CreateGroupEventForm groupId={groupId} onEventCreated={handleEventCreated} />
        </div>
      )}

      <GroupEventList
        events={events}
        groupId={groupId}
        onEventUpdated={handleEventUpdated}
        isLoading={eventsLoading}
        error={eventsError}
      />
      {/* Optional: Add a refresh button or pagination controls here */}
    </div>
  );
}