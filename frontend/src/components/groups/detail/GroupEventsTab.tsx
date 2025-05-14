import React from 'react';
import { Text } from '../../ui/text';
import { EventSummary } from '../../../types/Group';

interface GroupEventsTabProps {
  events: EventSummary[] | undefined;
}

export default function GroupEventsTab({ events }: GroupEventsTabProps) {
  if (!events || events.length === 0) {
    return <Text className="text-center text-gray-400 py-4">No events in this group yet.</Text>;
  }

  return (
    <div className="space-y-4">
      {events.map((event, index) => (
        // SIMPLIFIED RENDERING as per original component:
        <div key={event.id || index}>
          <Text>Event Title: {event.title || "No title"}</Text>
          <Text>Event ID: {event.id}</Text>
          {/* Do NOT try to render event.event_time here for now unless fully safeguarded */}
          <Text>Raw Event Time (from log): {event.event_time === undefined ? "undefined" : event.event_time || "empty/null"}</Text>
        </div>
      ))}
    </div>
  );
}