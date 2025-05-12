'use client';

import React, { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { UserSummary } from '@/types/User';
import { useRequest } from '@/hooks/useRequest';
import FollowRequestList from './FollowRequestList';
import { Tab } from '@headlessui/react'; // Using Headless UI for tabs

function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(' ');
}

interface FollowRequestsResponse {
  received: UserSummary[];
  sent: UserSummary[] | null; // API can return null for sent
}

const ManageFollowRequestsSection: React.FC = () => {
  const { get, post, del, isLoading, error } = useRequest<{ received: UserSummary[], sent: UserSummary[] | null }>();
  const [receivedRequests, setReceivedRequests] = useState<UserSummary[]>([]);
  const [sentRequests, setSentRequests] = useState<UserSummary[]>([]);
  const [loadingActionUserId, setLoadingActionUserId] = useState<string | null>(null);

  const fetchFollowRequests = useCallback(() => {
    get('/api/users/me/follow-requests', (data) => {
      if (data) {
        setReceivedRequests(data.received || []);
        setSentRequests(data.sent || []); // Handle null for sent
      }
    });
  }, [get]);

  useEffect(() => {
    fetchFollowRequests();
  }, [fetchFollowRequests]);

  // requesterId is the id of the user who sent the request
  const handleAccept = async (requesterId: string) => {
    setLoadingActionUserId(requesterId);
    try {
      await post(`/api/users/${requesterId}/accept`, {});
      toast.success('Follow request accepted!');
      fetchFollowRequests();
    } catch (e: any) {
      const errorMessage = e?.response?.data?.error || e.message || 'Failed to accept follow request.';
      toast.error(errorMessage);
      console.error('Accept request failed:', e);
    } finally {
      setLoadingActionUserId(null);
    }
  };

  // requesterId is the id of the user who sent the request
  const handleDecline = async (requesterId: string) => {
    setLoadingActionUserId(requesterId);
    try {
      await del(`/api/users/${requesterId}/reject`);
      toast.success('Follow request declined.');
      fetchFollowRequests();
    } catch (e: any) {
      const errorMessage = e?.response?.data?.error || e.message || 'Failed to decline follow request.';
      toast.error(errorMessage);
      console.error('Decline request failed:', e);
    } finally {
      setLoadingActionUserId(null);
    }
  };

  // targetId is the id of the user to whom the request was sent
  const handleCancel = async (targetId: string) => {
    setLoadingActionUserId(targetId);
    try {
      await del(`/api/users/${targetId}/cancel-follow-request`);
      toast.success('Follow request cancelled.');
      fetchFollowRequests();
    } catch (e: any) {
      const errorMessage = e?.response?.data?.error || e.message || 'Failed to cancel follow request.';
      toast.error(errorMessage);
      console.error('Cancel request failed:', e);
    } finally {
      setLoadingActionUserId(null);
    }
  };
  
  const isLoadingAction = (userId: string) => loadingActionUserId === userId;

  if (isLoading && receivedRequests.length === 0 && sentRequests.length === 0) {
    return <p className="text-gray-400 text-center py-10">Loading requests...</p>;
  }

  if (error) {
    return <p className="text-red-500 text-center py-10">Error loading follow requests: {error.message}</p>;
  }

  const tabs = [
    { name: 'Received Requests', count: receivedRequests.length },
    { name: 'Sent Requests', count: sentRequests.length },
  ];

  return (
    <div className="w-full max-w-2xl mx-auto px-2 py-8 sm:px-0">
      <Tab.Group>
        <Tab.List className="flex space-x-1 rounded-xl bg-gray-700 p-1">
          {tabs.map((tab) => (
            <Tab
              key={tab.name}
              className={({ selected }) =>
                classNames(
                  'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-purple-200',
                  'ring-white ring-opacity-60 ring-offset-2 ring-offset-purple-400 focus:outline-none focus:ring-2',
                  selected
                    ? 'bg-gray-900 shadow'
                    : 'text-gray-300 hover:bg-gray-800/[0.6] hover:text-white'
                )
              }
            >
              {tab.name} ({tab.count})
            </Tab>
          ))}
        </Tab.List>
        <Tab.Panels className="mt-2">
          <Tab.Panel
            className={classNames(
              'rounded-xl bg-gray-800 p-3',
              'ring-white ring-opacity-60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2'
            )}
          >
            <FollowRequestList
              users={receivedRequests}
              type="received"
              onAccept={handleAccept}
              onDecline={handleDecline}
              isLoadingAction={isLoadingAction}
              emptyMessage="No new follow requests."
            />
          </Tab.Panel>
          <Tab.Panel
            className={classNames(
              'rounded-xl bg-gray-800 p-3',
              'ring-white ring-opacity-60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2'
            )}
          >
            <FollowRequestList
              users={sentRequests}
              type="sent"
              onCancel={handleCancel}
              isLoadingAction={isLoadingAction}
              emptyMessage="You haven't sent any follow requests."
            />
          </Tab.Panel>
        </Tab.Panels>
      </Tab.Group>
    </div>
  );
};

export default ManageFollowRequestsSection;