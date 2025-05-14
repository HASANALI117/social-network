'use client';

import React from 'react';
import { GroupEvent } from '@/types/GroupEvent';
import { GroupEventCard } from './GroupEventCard';
import { Text } from '@/components/ui/text';

interface GroupEventListProps {
  events: GroupEvent[];
  groupId: string;
  onEventUpdated: (updatedEvent: GroupEvent) => void;
  isLoading?: boolean;
  error?: Error | null;
}

export function GroupEventList({
  events,
  groupId,
  onEventUpdated,
  isLoading,
  error,
}: GroupEventListProps) {
  if (isLoading) {
    return <Text>Loading events...</Text>;
  }

  if (error) {
    return <Text className="text-red-500">Error loading events: {error.message}</Text>;
  }

  // Check if events is a valid array and has items.
  // If not, or if it's empty, display a placeholder.
  if (!Array.isArray(events) || events.length === 0) {
    return <Text>No events scheduled for this group.</Text>;
  }

  return (
    <div className="space-y-6">
      {events.map((event) => (
        <GroupEventCard
          key={event.id}
          event={event}
          groupId={groupId}
          onEventUpdated={onEventUpdated}
        />
      ))}
    </div>
  );
}