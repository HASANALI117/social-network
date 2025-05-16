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
    // newEvent is of type GroupEvent, which doesn't have an 'options' property.
    // If 'options' were intended here, the GroupEvent type or the type of newEvent
    // would need to be adjusted. Assuming newEvent is a standard GroupEvent.
    setEvents((prevEvents) => [newEvent, ...prevEvents]);
  };

  const handleEventUpdated = (updatedEvent: GroupEvent) => {
    setEvents((prevEvents) =>
      prevEvents.map((event) => (event.id === updatedEvent.id ? updatedEvent : event))
    );
  };

  if (groupLoading) return (
    <div className="flex justify-center items-center h-64 min-h-screen bg-gray-900">
      <Text className="text-xl text-gray-400">Loading group details...</Text>
    </div>
  );
  if (groupError) return (
    <div className="min-h-screen bg-gray-900 text-gray-100 py-8 flex justify-center items-start">
      <div className="container mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
        <div className="bg-red-800 border border-red-700 text-red-200 px-4 py-3 rounded-lg shadow-md" role="alert">
          <Text className="font-bold">Error loading group:</Text>
          <Text className="block sm:inline"> {groupError.message}</Text>
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100 py-8">
      <div className="container mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 space-y-8">
        <Heading level={1} className="text-3xl sm:text-4xl font-bold text-white text-center sm:text-left">
          Events for {group ? `"${group.name}"` : 'Group'}
        </Heading>

        {groupId && (
          <div className="p-6 bg-gray-800 rounded-lg shadow-xl">
            <Heading level={2} className="text-2xl font-semibold text-white mb-4">
              Create New Event
            </Heading>
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
    </div>
  );
}