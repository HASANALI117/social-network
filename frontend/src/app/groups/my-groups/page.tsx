'use client';

import React, { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';
import { useRequest } from '../../../hooks/useRequest';
import { Group } from '../../../types/Group';
import { Button } from '../../../components/ui/button';
import { Heading } from '../../../components/ui/heading';
import { Text } from '../../../components/ui/text';
import GroupCard from '../../../components/groups/GroupCard';
import { useUserStore } from '../../../store/useUserStore';

interface MyGroupsApiResponse {
  groups: Group[];
}

export default function MyGroupsPage() {
  const [myGroups, setMyGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const { user } = useUserStore();
  const { get: fetchMyGroupsRequest, error: fetchGroupsError } = useRequest<MyGroupsApiResponse>();

  const loadMyGroups = useCallback(async () => {
    if (!user?.id) {
      setMyGroups([]);
      setIsLoading(false);
      setError(new Error("Please log in to see your groups."));
      return;
    }

    setIsLoading(true);
    setError(null);
    const url = '/api/users/me/groups';

    try {
      const data = await fetchMyGroupsRequest(url);
      if (data && data.groups) {
        setMyGroups(data.groups);
      } else if (fetchGroupsError) {
        setError(fetchGroupsError);
        setMyGroups([]);
      } else {
        setMyGroups([]);
        if (!data) setError(new Error("Failed to fetch your groups: No data returned."));
      }
    } catch (err: any) {
      setError(err);
      setMyGroups([]);
    } finally {
      setIsLoading(false);
    }
  }, [fetchMyGroupsRequest, fetchGroupsError, user?.id]);

  useEffect(() => {
    loadMyGroups();
  }, [loadMyGroups]);

   useEffect(() => {
    if (fetchGroupsError && error?.message !== fetchGroupsError.message) {
      setError(fetchGroupsError);
      setMyGroups([]);
      setIsLoading(false);
    }
  }, [fetchGroupsError, error]);


  const renderMyGroupList = () => {
    if (isLoading) return <Text className="text-center py-10 text-gray-400">Loading your groups...</Text>;
    if (error) return <Text className="text-center text-red-400 py-10">Error: {error.message}</Text>;
    if (!user?.id && !isLoading) return <Text className="text-center text-gray-400 py-10">Please log in to see your groups.</Text>;
    if (myGroups.length === 0) {
      return (
        <Text className="text-center text-gray-400 py-10">
          You are not a member of any groups yet.
        </Text>
      );
    }
    return (
      <div className="space-y-6">
        {myGroups.map((group) => (
          <GroupCard key={group.id} group={group} />
        ))}
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100 py-8">
      <div className="container mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col sm:flex-row justify-between items-center mb-8 gap-4">
          <Heading level={1} className="text-3xl sm:text-4xl font-bold text-white">
            My Groups
          </Heading>
          <Link href="/groups/create" passHref>
            <Button className="w-full sm:w-auto whitespace-nowrap bg-indigo-600 hover:bg-indigo-500 text-white">
              Create New Group
            </Button>
          </Link>
        </div>
        {renderMyGroupList()}
      </div>
    </div>
  );
}