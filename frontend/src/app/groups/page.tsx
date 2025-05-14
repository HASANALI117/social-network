'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import Image from 'next/image'; // For avatar
import { useRequest } from '@/hooks/useRequest';
import { Group } from '@/types/Group';
import { Button } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { Input } from '@/components/ui/input'; // For search bar
import { UserCircleIcon } from '@heroicons/react/24/solid'; // Placeholder icon

interface GroupsApiResponse {
  groups: Group[];
  count?: number; // Optional, as per instructions
  limit?: number; // Optional
  offset?: number; // Optional
}

// Debounce function
const debounce = <F extends (...args: any[]) => any>(func: F, waitFor: number) => {
  let timeout: ReturnType<typeof setTimeout> | null = null;

  const debounced = (...args: Parameters<F>) => {
    if (timeout !== null) {
      clearTimeout(timeout);
      timeout = null;
    }
    timeout = setTimeout(() => func(...args), waitFor);
  };

  return debounced as (...args: Parameters<F>) => ReturnType<F>;
};

export default function BrowseGroupsPage() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  const { get: fetchGroupsRequest, error: fetchGroupsError } = useRequest<GroupsApiResponse>();

  const loadGroups = useCallback(async (currentSearchTerm?: string) => {
    setIsLoading(true);
    setError(null);
    let url = '/api/groups';
    if (currentSearchTerm && currentSearchTerm.trim() !== '') {
      url += `?search=${encodeURIComponent(currentSearchTerm.trim())}`;
    }
    try {
      const data = await fetchGroupsRequest(url);
      if (data && data.groups) {
        setGroups(data.groups);
      } else if (fetchGroupsError) {
        setError(fetchGroupsError);
      } else {
        // If data.groups is not present but no explicit fetchGroupsError, it might be an empty successful response
        // or an unexpected API response structure.
        setGroups([]); // Assume no groups if not explicitly provided or error.
        if (!data) setError(new Error('Failed to fetch groups: No data returned.'));
      }
    } catch (err: any) {
      setError(err);
    } finally {
      setIsLoading(false);
    }
  }, [fetchGroupsRequest, fetchGroupsError]);

  const debouncedLoadGroups = useCallback(debounce(loadGroups, 400), [loadGroups]);

  useEffect(() => {
    // Initial load or when search term changes (debounced)
    debouncedLoadGroups(searchTerm);
  }, [searchTerm, debouncedLoadGroups]);


  useEffect(() => {
    // Handle direct fetch errors
    if (fetchGroupsError) {
      setError(fetchGroupsError);
      setGroups([]); // Clear groups on error
      setIsLoading(false);
    }
  }, [fetchGroupsError]);

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
  };

  const formatCreatorName = (creatorInfo: Group['creator_info']) => {
    if (creatorInfo.first_name && creatorInfo.last_name) {
      return `${creatorInfo.first_name} ${creatorInfo.last_name}`;
    }
    return creatorInfo.username;
  };

  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleDateString(undefined, {
        year: 'numeric', month: 'long', day: 'numeric'
      });
    } catch (e) {
      return "Invalid date";
    }
  };

  return (
    <div className="container mx-auto p-4 text-white">
      <div className="flex flex-col sm:flex-row justify-between items-center mb-8 gap-4">
        <Heading level={1} className="whitespace-nowrap">Browse Groups</Heading>
        <Input
          type="text"
          placeholder="Search groups..."
          value={searchTerm}
          onChange={handleSearchChange}
          className="w-full sm:w-auto bg-gray-700 border-gray-600 placeholder-gray-400 text-white"
        />
        <Link href="/groups/create" passHref>
          <Button className="w-full sm:w-auto whitespace-nowrap">Create New Group</Button>
        </Link>
      </div>

      {isLoading && <Text className="text-center py-10">Loading groups...</Text>}
      {error && <Text className="text-center text-red-500 py-10">Error: {error.message}</Text>}
      
      {!isLoading && !error && groups.length === 0 && (
        <Text className="text-center text-gray-400 py-10">
          {searchTerm ? `No groups found for "${searchTerm}".` : "No groups found. Be the first to create one!"}
        </Text>
      )}

      {!isLoading && !error && groups.length > 0 && (
        <div className=""> {/* Changed from grid to vertical stack with space */}
          {groups.map((group) => (
            <Link key={group.id} href={`/groups/${group.id}`} passHref>
              <div className="bg-gray-800 p-4 sm:p-6 rounded-lg shadow-lg hover:bg-gray-700 transition-colors cursor-pointer flex flex-col sm:flex-row gap-4 items-start mb-4">
                {group.avatar_url ? (
                  <Image src={group.avatar_url} alt={`${group.name} avatar`} width={80} height={80} className="rounded-md object-cover w-20 h-20 flex-shrink-0" />
                ) : (
                  <div className="w-20 h-20 bg-gray-700 rounded-md flex items-center justify-center flex-shrink-0">
                    <UserCircleIcon className="h-12 w-12 text-gray-500" />
                  </div>
                )}
                <div className="flex-grow">
                  <Heading level={3} className="mb-1 truncate">{group.name}</Heading>
                  <Text className="text-gray-400 text-sm line-clamp-3 mb-2">
                    {group.description || 'No description provided.'}
                  </Text>
                  <div className="text-xs text-gray-500 space-y-1">
                    <Text>Created by: {formatCreatorName(group.creator_info)}</Text>
                    <Text>Members: {group.members_count} | Posts: {group.posts_count} | Events: {group.events_count}</Text>
                    <Text>Created: {formatDate(group.created_at)}</Text>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
      {/* Add pagination controls here later if needed */}
    </div>
  );
}
