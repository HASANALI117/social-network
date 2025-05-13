'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useRequest } from '@/hooks/useRequest';
import { Group } from '@/types/Group';
import { Button } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';

interface GroupsApiResponse {
  groups: Group[];
  // Add pagination fields if your GET /api/groups endpoint supports them
}

export default function BrowseGroupsPage() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const { get: fetchGroupsRequest, error: fetchGroupsError } = useRequest<GroupsApiResponse>();

  const loadGroups = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await fetchGroupsRequest('/api/groups');
      if (data && data.groups) {
        setGroups(data.groups);
      } else if (fetchGroupsError) {
        setError(fetchGroupsError);
      } else {
        setError(new Error('Failed to fetch groups.'));
      }
    } catch (err: any) {
      setError(err);
    } finally {
      setIsLoading(false);
    }
  }, [fetchGroupsRequest, fetchGroupsError]);

  useEffect(() => {
    loadGroups();
  }, [loadGroups]);

  useEffect(() => {
    if (fetchGroupsError) {
      setError(fetchGroupsError);
      setIsLoading(false); // Ensure loading stops on error
    }
  }, [fetchGroupsError]);

  return (
    <div className="container mx-auto p-4 text-white">
      <div className="flex justify-between items-center mb-8">
        <Heading level={1}>Browse Groups</Heading>
        <Link href="/groups/create" passHref>
          <Button>Create New Group</Button>
        </Link>
      </div>

      {isLoading && <Text className="text-center">Loading groups...</Text>}
      {error && <Text className="text-center text-red-500">Error: {error.message}</Text>}
      
      {!isLoading && !error && groups.length === 0 && (
        <Text className="text-center text-gray-400">No groups found. Be the first to create one!</Text>
      )}

      {!isLoading && !error && groups.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {groups.map((group) => (
            <Link key={group.id} href={`/groups/${group.id}`} passHref>
              <div className="bg-gray-800 p-6 rounded-lg shadow-lg hover:bg-gray-700 transition-colors cursor-pointer">
                <Heading level={3} className="mb-2 truncate">{group.title}</Heading>
                <Text className="text-gray-400 text-sm line-clamp-3 mb-1">
                  {group.description || 'No description provided.'}
                </Text>
                <Text className="text-xs text-gray-500 mt-3">
                  Created by: User {group.creator_id.substring(0,8)}...
                </Text>
                 {/* We can add member count or creator name later if backend provides it */}
              </div>
            </Link>
          ))}
        </div>
      )}
      {/* Add pagination controls here later if needed */}
    </div>
  );
}
