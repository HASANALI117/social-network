'use client';
import { useEffect, useState } from 'react';
import { useRequest } from '@/hooks/useRequest'; // Import ApiError
import { useUserStore } from '@/store/useUserStore';
import { GroupJoinRequest } from '@/types/GroupJoinRequest';
import { UserBasicInfo } from '@/types/User';
import { Avatar } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Text } from '@/components/ui/text';
import { Alert } from '@/components/ui/alert';
import { Heading } from '@/components/ui/heading';
import toast from 'react-hot-toast';

interface GroupJoinRequestsTabProps {
  groupId: string;
}

export default function GroupJoinRequestsTab({ groupId }: GroupJoinRequestsTabProps) {
  const [pendingJoinRequests, setPendingJoinRequests] = useState<GroupJoinRequest[]>([]);
  const [processingRequestId, setProcessingRequestId] = useState<string | null>(null);
  const { isAuthenticated, hydrated } = useUserStore();

  const {
    isLoading,
    error: fetchError,
    get: fetchPendingJoinRequests
  } = useRequest<{ count: number; requests: GroupJoinRequest[] }>();
  
  const acceptRequestHook = useRequest<void>();
  const rejectRequestHook = useRequest<void>();

  useEffect(() => {
    // Only fetch if user is authenticated, store is hydrated, and groupId is provided
    if (groupId && isAuthenticated && hydrated) {
      fetchPendingJoinRequests(
        `/api/groups/${groupId}/requests/pending`,
        (apiResponse) => { // This is the onSuccess callback
          console.log('GroupJoinRequestsTab: Raw API response (onSuccess):', JSON.stringify(apiResponse, null, 2));
          if (apiResponse && Array.isArray(apiResponse.requests)) {
            setPendingJoinRequests(apiResponse.requests);
            console.log('GroupJoinRequestsTab: Set pendingJoinRequests (onSuccess):', apiResponse.requests);
          } else {
            console.warn('GroupJoinRequestsTab: API response structure incorrect or .requests missing (onSuccess). Response:', apiResponse);
            setPendingJoinRequests([]);
          }
        }
      );
    } else if (!isAuthenticated) {
      // Clear data if user is not authenticated
      setPendingJoinRequests([]);
    }
  }, [groupId, fetchPendingJoinRequests, isAuthenticated, hydrated]);

  const handleAccept = async (requestId: string) => {
    setProcessingRequestId(requestId);
    const result = await acceptRequestHook.post(`/api/groups/requests/${requestId}/accept`, {});
    if (result !== null && !acceptRequestHook.error) {
      setPendingJoinRequests((prevRequests) =>
        prevRequests.filter((req) => req.id !== requestId)
      );
      toast.success('Join request accepted.');
    } else {
      toast.error(acceptRequestHook.error?.message || 'Failed to accept join request.');
    }
    setProcessingRequestId(null);
  };

  const handleReject = async (requestId: string) => {
    setProcessingRequestId(requestId);
    const result = await rejectRequestHook.post(`/api/groups/requests/${requestId}/reject`, {});
    if (result !== null && !rejectRequestHook.error) {
      setPendingJoinRequests((prevRequests) =>
        prevRequests.filter((req) => req.id !== requestId)
      );
      toast.success('Join request rejected.');
    } else {
      toast.error(rejectRequestHook.error?.message || 'Failed to reject join request.');
    }
    setProcessingRequestId(null);
  };

  if (isLoading) {
    return <Text>Loading join requests...</Text>;
  }

  if (fetchError) {
    // Log the error for debugging if it's not a 403 or not an ApiError with status
    console.error("Error loading join requests:", fetchError);
    return <Text className="text-red-500">Error loading join requests: {fetchError.message}</Text>;
  }

  return (
    <div className="space-y-4">
      <Heading level={3}>Pending Join Requests</Heading>
      {!isLoading && !fetchError && Array.isArray(pendingJoinRequests) && pendingJoinRequests.length > 0 && (
        pendingJoinRequests.map((request) => {
          const requester = request.requester;
          if (!requester) return null;

          return (
            <div
              key={request.id}
              className="flex items-center justify-between p-4 border rounded-md"
            >
              <div className="flex items-center space-x-3">
                <Avatar
                  src={requester.avatar_url || undefined}
                  alt={requester.username}
                  initials={requester.username.charAt(0).toUpperCase()}
                  className="h-10 w-10 rounded-full mr-3"
                />
                <div>
                  <Text className="font-medium">
                    {requester.first_name} {requester.last_name}
                  </Text>
                  <Text className="text-sm text-gray-600">
                    @{requester.username}
                  </Text>
                </div>
              </div>
              <div className="flex space-x-2">
                <Button
                  onClick={() => handleAccept(request.id)}
                  disabled={processingRequestId === request.id || acceptRequestHook.isLoading || rejectRequestHook.isLoading}
                  outline={true}
                  className="text-green-600 border-green-600 hover:bg-green-50 dark:text-green-400 dark:border-green-500 dark:hover:bg-green-900/30" // Custom styling for green outline
                >
                  {processingRequestId === request.id && acceptRequestHook.isLoading ? 'Accepting...' : 'Accept'}
                </Button>
                <Button
                  onClick={() => handleReject(request.id)}
                  disabled={processingRequestId === request.id || acceptRequestHook.isLoading || rejectRequestHook.isLoading}
                  outline={true}
                  className="text-red-600 border-red-600 hover:bg-red-50 dark:text-red-400 dark:border-red-500 dark:hover:bg-red-900/30" // Custom styling for red outline
                >
                  {processingRequestId === request.id && rejectRequestHook.isLoading ? 'Rejecting...' : 'Reject'}
                </Button>
              </div>
            </div>
          );
        })
      )}
      {!isLoading && !fetchError && (!Array.isArray(pendingJoinRequests) || pendingJoinRequests.length === 0) && (
        <Text>No pending join requests.</Text>
      )}
    </div>
  );
}