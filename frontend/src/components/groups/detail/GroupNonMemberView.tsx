import React, { useState, useEffect } from 'react';
import { Heading } from '../../ui/heading';
import { Text } from '../../ui/text';
import { Button } from '../../ui/button';
import { Group } from '@/types/Group'; // Assuming Group is the detailed type
import { useRequest } from '@/hooks/useRequest';
import { GroupJoinRequest } from '@/types/GroupJoinRequest';
import toast from 'react-hot-toast';
// Removed Alert import as we will rely on toast for feedback

interface GroupNonMemberViewProps {
  group: Group;
}

export default function GroupNonMemberView({ group }: GroupNonMemberViewProps) {
  const [isSubmittingRequest, setIsSubmittingRequest] = useState(false);
  const [requestError, setRequestError] = useState<string | null>(null);
  const [requestSuccessMessage, setRequestSuccessMessage] = useState<string | null>(null);
  const [hasPendingRequest, setHasPendingRequest] = useState(false);

  const { post: submitJoinRequest, isLoading: apiIsLoading, error: apiError } = useRequest<GroupJoinRequest>();

  useEffect(() => {
    if (group?.viewer_pending_request_status === 'pending') {
      setHasPendingRequest(true);
    } else {
      setHasPendingRequest(false); // Reset if status changes
    }
  }, [group?.viewer_pending_request_status]);

  const handleRequestToJoin = async () => {
    setIsSubmittingRequest(true);
    setRequestError(null);
    setRequestSuccessMessage(null);

    try {
      const response = await submitJoinRequest(`/api/groups/${group.id}/requests`, {});
      if (response) {
        setHasPendingRequest(true);
        const successMsg = "Your request to join has been sent.";
        setRequestSuccessMessage(successMsg);
        toast.success(successMsg);
      } else if (apiError) {
        // Error handled by the effect below
      }
    } catch (error) {
      // This catch might be redundant if useRequest handles all errors via apiError
      const errorMsg = "An unexpected error occurred. Please try again.";
      setRequestError(errorMsg);
      toast.error(errorMsg);
    } finally {
      setIsSubmittingRequest(false);
    }
  };

  useEffect(() => {
    if (apiError) {
      const errorMsg = apiError.message || "Failed to send join request.";
      setRequestError(errorMsg);
      toast.error(errorMsg);
    }
  }, [apiError]);


  const isLoading = isSubmittingRequest || apiIsLoading;
  let buttonText = "Request to Join";
  let buttonDisabled = false;

  if (hasPendingRequest) {
    buttonText = "Pending Request";
    buttonDisabled = true;
  } else if (isLoading) {
    buttonText = "Sending Request...";
    buttonDisabled = true;
  }

  return (
    <div className="text-center">
      <Heading level={2} className="text-2xl mb-4 text-gray-200">
        Join the Conversation!
      </Heading>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6 text-lg">
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{group.members_count}</Text>
          <Text className="text-gray-300">Members</Text>
        </div>
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{group.posts_count}</Text>
          <Text className="text-gray-300">Posts</Text>
        </div>
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{group.events_count}</Text>
          <Text className="text-gray-300">Events</Text>
        </div>
      </div>

      {/* requestError and requestSuccessMessage are handled by toast notifications */}

      <Button
        onClick={handleRequestToJoin}
        disabled={buttonDisabled}
        className="bg-purple-600 hover:bg-purple-700 text-white font-bold py-3 px-6 rounded-lg text-lg disabled:opacity-50"
      >
        {buttonText}
      </Button>
    </div>
  );
}