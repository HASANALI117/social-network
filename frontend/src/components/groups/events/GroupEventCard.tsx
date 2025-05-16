'use client';

import React from 'react';
import Link from 'next/link'; // Added import
import { GroupEvent } from '@/types/GroupEvent';
// GroupEventResponseOption might be part of GroupEvent or deprecated
// For now, assuming options are part of GroupEvent or handled differently
import { useRequest } from '@/hooks/useRequest';
import { Button } from '@/components/ui/button';
import { Text }
from '@/components/ui/text';
import { Heading } from '@/components/ui/heading';
import { useUserStore } from '@/store/useUserStore'; // To get current user ID

interface GroupEventCardProps {
  event: GroupEvent;
  groupId: string;
  onEventUpdated: (updatedEvent: GroupEvent) => void;
}

export function GroupEventCard({ event, groupId, onEventUpdated }: GroupEventCardProps) {
  const { user } = useUserStore();
  const { post, isLoading, error } = useRequest<GroupEvent>();

  const handleResponse = async (optionId: string) => {
    if (!user) {
      // Handle case where user is not logged in, though ideally this component wouldn't be shown
      alert('You must be logged in to respond.');
      return;
    }

    await post(
      `/api/groups/${groupId}/events/${event.id}/responses`,
      { option_id: optionId },
      (updatedEvent) => {
        onEventUpdated(updatedEvent);
      }
    );
  };

  const formatDateTime = (isoString: string) => {
    if (!isoString) return 'Date not set';
    try {
      return new Date(isoString).toLocaleString(undefined, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch (e) {
      return 'Invalid Date';
    }
  };

  const creatorName = event.creator_name || event.creator_id;

  return (
    <Link href={`/groups/${event.group_id}/events/${event.id}`} passHref legacyBehavior={false}>
      <div className="border rounded-lg p-4 shadow-sm mb-4 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer transition-colors duration-150">
        <Heading level={4} className="mb-2">{event.title}</Heading>
        <Text className="text-gray-700 dark:text-gray-300 mb-1">{event.description}</Text>
        <Text className="text-sm text-gray-500 dark:text-gray-400 mb-1">
          Time: {formatDateTime(event.event_time)}
        </Text>
        <Text className="text-sm text-gray-500 dark:text-gray-400 mb-3">
          Created by: {creatorName}
        </Text>

        <div className="flex space-x-2 mb-3">
          {/* This section needs to be updated based on how responses are now handled.
              The GroupEvent type might not have 'options' or 'current_user_response_id' directly in the list view.
              For simplicity, I'll comment this out as the card is primarily for navigation.
              Response interaction is detailed on the event detail page.
          */}
          {/* {Array.isArray(event.options) && event.options.length > 0 ? (
            event.options.map((option: any) => ( // Replace 'any' with actual type if options are still used
              <Button
                key={option.id}
                onClick={(e: React.MouseEvent) => {
                  e.stopPropagation();
                  handleResponse(option.id);
                }}
                disabled={isLoading} // Simplified disabling logic
                // className logic would also need an update
              >
                {option.text} ({option.count})
              </Button>
            ))
          ) : (
            <Text className="text-sm text-gray-500 dark:text-gray-400">Response options not shown on card.</Text>
          )} */}
          <Text className="text-sm text-gray-500 dark:text-gray-400">
            Click to view details and respond. Responses: {event.response_count ?? 0}
          </Text>
          {/* The ternary operator causing the syntax error was here.
              It seems like an artifact from the previous structure.
              The `Text` component above is now the sole content of this div.
           */}
        </div>

        {error && (
          <Text className="text-red-500 text-sm mt-2">
            Error responding: {error.message}
          </Text>
        )}
      </div>
    </Link>
  );
}