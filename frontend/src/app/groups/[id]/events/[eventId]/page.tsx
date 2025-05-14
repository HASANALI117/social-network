'use client';

import { useParams } from 'next/navigation';
import { useRequest } from '@/hooks/useRequest';
import { GroupEvent, GroupEventResponseOption } from '@/types/GroupEvent';
import UserCard from '@/components/profile/UserCard'; // Default import
import UserList from '@/components/profile/UserList'; // Default import
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { Button } from '@/components/ui/button';
// Using Text for simpler inline errors instead of modal Alert for now
// import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { useEffect, useState, useCallback } from 'react';
import { useUserStore } from '@/store/useUserStore';
import { User, UserBasicInfo } from '@/types/User';

interface GroupEventDetailPageProps {
  params: {
    id: string;
    eventId: string;
  };
}

export default function GroupEventDetailPage({ params }: GroupEventDetailPageProps) {
  const { id: groupId, eventId } = params;
  const { user } = useUserStore();
  const [optimisticEvent, setOptimisticEvent] = useState<GroupEvent | null>(null);

  // useRequest for fetching event details
  const {
    get: fetchEvent,
    data: eventData,
    error: fetchError,
    isLoading: isLoadingEvent
  } = useRequest<GroupEvent>();

  // useRequest for submitting response
  const {
    post: postResponse,
    error: responseError,
    isLoading: isResponding
  } = useRequest<GroupEvent>();

  const loadEventDetails = useCallback(async () => {
    if (groupId && eventId) {
      const fetchedEvent = await fetchEvent(`/api/groups/${groupId}/events/${eventId}`);
      if (fetchedEvent) {
        setOptimisticEvent(fetchedEvent);
      }
    }
  }, [groupId, eventId, fetchEvent]);

  useEffect(() => {
    loadEventDetails();
  }, [loadEventDetails]);
  
  // Update optimistic event if eventData from initial fetch changes (e.g. after a successful post and re-fetch)
  useEffect(() => {
    if (eventData) {
      setOptimisticEvent(eventData);
    }
  }, [eventData]);

  const handleResponse = async (optionId: string) => {
    if (!user || !optimisticEvent) return;

    const originalEvent = JSON.parse(JSON.stringify(optimisticEvent)) as GroupEvent; // For rollback
    let tempEvent = JSON.parse(JSON.stringify(optimisticEvent)) as GroupEvent;

    const currentOptionId = tempEvent.current_user_response_id;
    
    // Update counts and user lists optimistically
    tempEvent.options = tempEvent.options.map(opt => {
      let newUsers = opt.users ? [...opt.users] : [];
      let newCount = opt.count;

      if (opt.id === currentOptionId) { // User is changing from this option
        newCount = Math.max(0, newCount - 1);
        newUsers = newUsers.filter(u => u.user_id !== user.id);
      }
      if (opt.id === optionId) { // User is choosing this new option
        if (currentOptionId !== optionId) { // Only add if not already part of it (or if it's a new response)
            newCount += 1;
            // Ensure user is not duplicated if somehow already in list
            if (!newUsers.find(u => u.user_id === user.id)) {
                 newUsers.push({ user_id: user.id, username: user.username, first_name: user.first_name, last_name: user.last_name, avatar_url: user.avatar_url });
            }
        }
      }
      return { ...opt, count: newCount, users: newUsers };
    });
    
    tempEvent.current_user_response_id = optionId;
    setOptimisticEvent(tempEvent);

    try {
      const updatedDataFromServer = await postResponse(
        `/api/groups/${groupId}/events/${eventId}/responses`,
        { option_id: optionId }
      );
      if (updatedDataFromServer) {
        setOptimisticEvent(updatedDataFromServer); // Sync with server state
      } else if (responseError) { // Explicitly check for error from postResponse
        setOptimisticEvent(originalEvent); // Rollback on error
        console.error("Failed to submit response:", responseError);
      }
    } catch (err) { // Catch any other errors during post
      setOptimisticEvent(originalEvent); // Rollback
      console.error("Failed to submit response:", err);
    }
  };
  
  // Helper to map UserBasicInfo to User for UserCard/UserList
  // UserCard/UserList expect `id` not `user_id`.
  // User type has more fields, but UserCard primarily uses these.
  const mapBasicInfoToUser = (basicInfo: UserBasicInfo): User => ({
    id: basicInfo.user_id, // Map user_id to id
    username: basicInfo.username,
    first_name: basicInfo.first_name,
    last_name: basicInfo.last_name,
    avatar_url: basicInfo.avatar_url,
    // Add dummy values for other required User fields if necessary for UserCard/UserList
    // UserCard primarily uses id, username, first_name, last_name, avatar_url
    email: '',
    birth_date: '',
    is_private: false,
    created_at: '',
    updated_at: '',
    // followers_count and following_count are not on the base User type
    // and not strictly needed by UserCard as per its implementation.
  });


  if (isLoadingEvent && !optimisticEvent) return <Text>Loading event details...</Text>;
  if (fetchError) return <Text className="text-red-500">Error loading event: {fetchError.message}</Text>;
  if (!optimisticEvent) return <Text>Event not found.</Text>;

  const { title, description, event_time, creator_info, options } = optimisticEvent;

  return (
    <div className="container mx-auto p-4 space-y-6">
      <Heading level={1} className="mb-4 text-3xl font-bold">{title}</Heading>
      
      <div className="bg-white shadow-md rounded-lg p-6">
        <Text className="text-gray-700 mb-4">{description}</Text>
        <Text className="text-sm text-gray-500 mb-2">
          Date & Time: {new Date(event_time).toLocaleString()}
        </Text>
        <div className="mb-4">
          <Text className="font-semibold mb-1">Created by:</Text>
          {creator_info && <UserCard user={mapBasicInfoToUser(creator_info)} />}
        </div>
      </div>

      <div className="space-y-2">
        <Text className="font-semibold">Your Response:</Text>
        <div className="flex space-x-2">
          {options.map(option => {
            const isSelected = optimisticEvent.current_user_response_id === option.id;
            return (
              <Button
                key={option.id}
                onClick={() => handleResponse(option.id)}
                {...(isSelected ? { color: 'blue' as const } : { outline: true as const })}
                disabled={isResponding || isLoadingEvent}
              >
                {option.text}
              </Button>
            );
          })}
        </div>
        {responseError && <Text className="text-red-500 mt-2">Failed to submit response: {responseError.message}</Text>}
      </div>

      <div className="space-y-4">
        {options.map(option => (
          <div key={option.id} className="bg-gray-50 p-4 rounded-lg">
            <Heading level={3} className="mb-2 text-xl font-semibold">{option.text} ({option.count})</Heading>
            {option.users && option.users.length > 0 ? (
              <UserList users={option.users.map(mapBasicInfoToUser)} />
            ) : (
              <Text className="text-sm text-gray-500">No one has responded with '{option.text}' yet.</Text>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}