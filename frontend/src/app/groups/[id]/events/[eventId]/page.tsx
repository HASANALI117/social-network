'use client';

import { useParams } from 'next/navigation';
import { useRequest } from '@/hooks/useRequest';
import { GroupEventDetail, IndividualEventResponse } from '@/types/GroupEvent';
import UserCard from '@/components/profile/UserCard'; // Default import
import UserList from '@/components/profile/UserList'; // Default import
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { Button } from '@/components/ui/button';
// Using Text for simpler inline errors instead of modal Alert for now
// import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import React, { useEffect, useState, useCallback } from 'react'; // Removed use
import { useUserStore } from '@/store/useUserStore';
import { User, UserBasicInfo } from '@/types/User';

// GroupEventDetailPageProps removed as params are now from useParams() via React.use()
export default function GroupEventDetailPage() { // Removed params from props
  // Define expected params structure
  interface PageRouteParams {
    id: string;
    eventId: string;
    [key: string]: string; // Add index signature
  }
  // Get params directly from useParams with the defined structure
  const params = useParams<PageRouteParams>();
  const groupId = params.id; // Now correctly typed from PageRouteParams
  const eventId = params.eventId; // Now correctly typed from PageRouteParams
  const { user } = useUserStore();
  const [optimisticEvent, setOptimisticEvent] = useState<GroupEventDetail | null>(null);

  // useRequest for fetching event details
  const {
    get: fetchEvent,
    data: eventData,
    error: fetchError,
    isLoading: isLoadingEvent
  } = useRequest<GroupEventDetail>();

  // useRequest for submitting response
  const {
    post: postResponse,
    error: responseError,
    isLoading: isResponding
  } = useRequest<GroupEventDetail>();

  const loadEventDetails = useCallback(async () => {
    if (groupId && eventId) {
      const fetchedEvent = await fetchEvent(`/api/groups/${groupId}/events/${eventId}`);
      if (fetchedEvent) {
        // Ensure responses is an array
        fetchedEvent.responses = Array.isArray(fetchedEvent.responses) ? fetchedEvent.responses : [];
        // Initialize response_counts if not present or not an object
        fetchedEvent.response_counts = typeof fetchedEvent.response_counts === 'object' && fetchedEvent.response_counts !== null
            ? fetchedEvent.response_counts
            : { going: 0, not_going: 0 }; // Default if not provided
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

  const handleResponse = async (newOptionKey: 'going' | 'not_going') => {
    if (!user || !optimisticEvent) return;

    const originalEvent = JSON.parse(JSON.stringify(optimisticEvent)) as GroupEventDetail; // For rollback
    let tempEvent = JSON.parse(JSON.stringify(optimisticEvent)) as GroupEventDetail; // For optimistic update

    const oldOptionKey = tempEvent.current_user_response_option_id;

    // If user clicks the same option they already selected, do nothing
    if (oldOptionKey === newOptionKey) {
        return;
    }

    // Optimistically update response_counts
    if (oldOptionKey) {
        tempEvent.response_counts[oldOptionKey] = Math.max(0, (tempEvent.response_counts[oldOptionKey] || 0) - 1);
    }
    tempEvent.response_counts[newOptionKey] = (tempEvent.response_counts[newOptionKey] || 0) + 1;

    // Optimistically update responses array
    const newUserResponse: IndividualEventResponse = {
        user_id: user.id,
        username: user.username,
        first_name: user.first_name,
        last_name: user.last_name,
        avatar_url: user.avatar_url,
        response: newOptionKey,
        updated_at: new Date().toISOString(),
    };
    tempEvent.responses = tempEvent.responses.filter(resp => resp.user_id !== user.id);
    tempEvent.responses.push(newUserResponse);

    // Optimistically update current user's response
    tempEvent.current_user_response_option_id = newOptionKey;
    
    setOptimisticEvent(tempEvent); // Apply optimistic update to UI

    try {
      // Submit the response to the server
      const submissionResult = await postResponse(
        `/api/groups/${groupId}/events/${eventId}/responses`,
        { response: newOptionKey }
      );

      if (submissionResult) {
        // If submission is successful, reload event details to get the canonical state from server
        // This ensures UI consistency if the POST response itself is minimal or doesn't return the full event detail.
        await loadEventDetails();
      } else if (responseError) {
        // If postResponse hook indicates an error (e.g., network issue before server responds, or non-2xx status handled by hook)
        console.error("Failed to submit response (hook error), rolling back:", responseError);
        setOptimisticEvent(originalEvent);
      }
      // Note: If postResponse throws an error (e.g. network failure), it will be caught by the catch block.
    } catch (err) {
      // Catch any error during submission or subsequent loadEventDetails
      console.error("Error during response submission or re-fetch, rolling back:", err);
      setOptimisticEvent(originalEvent);
    }
  };
  
  // Helper to map IndividualEventResponse (which now contains direct user fields) to User for UserCard/UserList
  const mapIndividualResponseToUser = (responseItem: IndividualEventResponse): User => ({
    id: responseItem.user_id, // Map user_id to id
    username: responseItem.username,
    first_name: responseItem.first_name || '', // Provide default if optional
    last_name: responseItem.last_name || '',   // Provide default if optional
    avatar_url: responseItem.avatar_url || undefined,
    // Add dummy values for other required User fields if necessary
    email: '',
    birth_date: '',
    is_private: false,
    created_at: '', // This could be responseItem.updated_at if relevant
    updated_at: responseItem.updated_at,
  });


  if (isLoadingEvent && !optimisticEvent) return <div className="min-h-screen bg-gray-900 text-gray-100 flex items-center justify-center"><Text className="text-gray-300 text-xl">Loading event details...</Text></div>;
  if (fetchError) return <div className="min-h-screen bg-gray-900 text-gray-100 flex items-center justify-center"><Text className="text-red-400 text-xl">Error loading event: {fetchError.message}</Text></div>;
  if (!optimisticEvent) return <div className="min-h-screen bg-gray-900 text-gray-100 flex items-center justify-center"><Text className="text-gray-400 text-xl">Event not found.</Text></div>;

  const { title, description, event_time, creator_info, creator_name, responses = [], response_counts = {}, current_user_response_option_id } = optimisticEvent;

  const EVENT_RESPONSE_OPTIONS = [
    { id: 'going' as const, text: 'Going' },
    { id: 'not_going' as const, text: 'Not Going' },
  ];

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100 py-8">
      <div className="container mx-auto max-w-3xl p-4 space-y-8">
        {/* Event Info Card */}
        <div className="bg-gray-800 shadow-xl rounded-lg p-6 space-y-4">
          <Heading level={1} className="text-3xl sm:text-4xl font-bold text-white mb-2">{title}</Heading>
          <Text className="text-gray-300 text-lg">{description}</Text>
          <Text className="text-sm text-indigo-400">
            Date & Time: {new Date(event_time).toLocaleString()}
          </Text>
          <div>
            <Text className="font-semibold text-gray-400 mb-1">Organized by:</Text>
            {creator_info ? (
              <UserCard user={mapIndividualResponseToUser(creator_info as unknown as IndividualEventResponse)} />
            ) : (
              <Text className="text-gray-300">{creator_name || optimisticEvent.creator_id || 'Unknown Creator'}</Text>
            )}
          </div>
        </div>

        {/* Response Buttons Section */}
        <div className="bg-gray-800 shadow-lg rounded-lg p-6 space-y-3">
          <Text className="font-semibold text-xl text-white">Your Response:</Text>
          <div className="flex space-x-4">
            {EVENT_RESPONSE_OPTIONS.map(option => {
              const isSelected = current_user_response_option_id === option.id;
              return (
                <Button
                  key={option.id}
                  onClick={() => handleResponse(option.id as 'going' | 'not_going')}
                  disabled={isResponding || isLoadingEvent}
                  className={`px-4 py-2 rounded-md font-medium transition-colors
                    ${isSelected
                      ? 'bg-indigo-600 hover:bg-indigo-700 text-white'
                      : 'bg-gray-700 hover:bg-gray-600 text-gray-200 hover:text-white'
                    }
                    ${isResponding || isLoadingEvent ? 'opacity-50 cursor-not-allowed' : ''}
                  `}
                >
                  {option.text}
                </Button>
              );
            })}
          </div>
          {responseError && <Text className="text-red-400 mt-2">Failed to submit response: {responseError.message}</Text>}
        </div>

        {/* Attendee Lists */}
        <div className="space-y-6">
          {EVENT_RESPONSE_OPTIONS.map(displayOption => {
            const optionKey = displayOption.id;
            const count = response_counts[optionKey] || 0;
            const usersForOption = responses.filter(r => r.response === optionKey);
            
            return (
              <div key={optionKey} className="bg-gray-800 shadow-lg rounded-lg p-6">
                <Heading level={3} className="text-2xl font-semibold text-white mb-4">{displayOption.text} ({count})</Heading>
                {usersForOption.length > 0 ? (
                  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                    {usersForOption.map(r => (
                      // Assuming UserCard has a dark theme-compatible styling or we adjust it here/globally
                      // For now, let's assume UserCard is somewhat neutral or we'll address it if it clashes.
                      // If UserCard needs explicit dark theme bg, it should be bg-gray-700 or similar.
                      <UserCard key={r.user_id} user={mapIndividualResponseToUser(r)} className="bg-gray-700 p-3 rounded-md" />
                    ))}
                  </div>
                ) : (
                  <Text className="text-gray-400 italic">No one has responded with '{displayOption.text}' yet.</Text>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}