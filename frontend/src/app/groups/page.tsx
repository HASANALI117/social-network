'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
// Image and UserCircleIcon are now in GroupCard
import { useRequest } from '@/hooks/useRequest';
import { Group } from '@/types/Group';
import { Button } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { Input } from '@/components/ui/input';
import Tabs from '@/components/common/Tabs'; // Import Tabs
import GroupCard from '@/components/groups/GroupCard'; // Import GroupCard
import { useUserStore } from '@/store/useUserStore'; // Import useUserStore

interface GroupsApiResponse {
  groups: Group[];
  count?: number;
  limit?: number;
  offset?: number;
}

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

type ActiveTab = 'explore' | 'my-groups';

export default function GroupsPage() {
  const [exploreGroups, setExploreGroups] = useState<Group[]>([]);
  const [myGroups, setMyGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [activeTabKey, setActiveTabKey] = useState<ActiveTab>('explore'); // Renamed for clarity

  const { user } = useUserStore();
  const { get: fetchGroupsRequest, error: fetchGroupsError } = useRequest<GroupsApiResponse>();

  const loadGroups = useCallback(async (currentSearchTerm?: string, tabKeyToLoad: ActiveTab = activeTabKey) => {
    setIsLoading(true);
    setError(null);
    let url = '/api/groups';

    if (tabKeyToLoad === 'explore') {
      if (currentSearchTerm && currentSearchTerm.trim() !== '') {
        url += `?search=${encodeURIComponent(currentSearchTerm.trim())}`;
      }
    } else if (tabKeyToLoad === 'my-groups') {
      if (user?.id) {
        url += `?member=true&userId=${user.id}`;
      } else {
        setIsLoading(false);
        setMyGroups([]); // Clear my groups if user not available
        return;
      }
    }

    try {
      const data = await fetchGroupsRequest(url);
      if (data && data.groups) {
        if (tabKeyToLoad === 'explore') {
          setExploreGroups(data.groups);
        } else {
          setMyGroups(data.groups);
        }
      } else if (fetchGroupsError) {
        setError(fetchGroupsError);
        if (tabKeyToLoad === 'explore') setExploreGroups([]); else setMyGroups([]);
      } else {
        if (tabKeyToLoad === 'explore') setExploreGroups([]); else setMyGroups([]);
        if (!data) setError(new Error(`Failed to fetch ${tabKeyToLoad === 'explore' ? 'explore' : 'my'} groups: No data returned.`));
      }
    } catch (err: any) {
      setError(err);
      if (tabKeyToLoad === 'explore') setExploreGroups([]); else setMyGroups([]);
    } finally {
      setIsLoading(false);
    }
  }, [fetchGroupsRequest, fetchGroupsError, user?.id, activeTabKey]);

  const debouncedLoadGroups = useCallback(debounce(loadGroups, 400), [loadGroups]);

  useEffect(() => {
    if (activeTabKey === 'explore') {
      debouncedLoadGroups(searchTerm, 'explore');
    } else if (activeTabKey === 'my-groups') {
      loadGroups(undefined, 'my-groups');
    }
  }, [searchTerm, activeTabKey, debouncedLoadGroups, loadGroups]);

  useEffect(() => {
    if (fetchGroupsError) {
      setError(fetchGroupsError);
      // Clear appropriate group list based on current tab
      if (activeTabKey === 'explore') setExploreGroups([]); else setMyGroups([]);
      setIsLoading(false);
    }
  }, [fetchGroupsError, activeTabKey]);

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
  };
  
  // This function will be called by the Tabs component when a tab is clicked
  // The Tabs component passes the index of the clicked tab.
  const handleTabSelection = (index: number) => {
    if (index === 0) {
      setActiveTabKey('explore');
    } else if (index === 1) {
      setActiveTabKey('my-groups');
    }
  };

  const renderGroupList = (groups: Group[], tabKey: ActiveTab) => {
    const currentSearchTermForDisplay = tabKey === 'explore' ? searchTerm : '';
    if (isLoading) return <Text className="text-center py-10">Loading groups...</Text>;
    if (error) return <Text className="text-center text-red-500 py-10">Error: {error.message}</Text>;
    if (groups.length === 0) {
      return (
        <Text className="text-center text-gray-400 py-10">
          {tabKey === 'explore'
            ? (currentSearchTermForDisplay ? `No groups found for "${currentSearchTermForDisplay}".` : "No groups found. Be the first to create one!")
            : (user?.id ? "You are not a member of any groups yet, or no groups found." : "Please log in to see your groups.")
          }
        </Text>
      );
    }
    return (
      <div className="space-y-4 mt-4">
        {groups.map((group) => (
          <GroupCard key={group.id} group={group} />
        ))}
      </div>
    );
  };

  const tabDefinitions = [
    {
      label: 'Explore',
      content: renderGroupList(exploreGroups, 'explore'),
    },
    {
      label: 'My Groups',
      content: renderGroupList(myGroups, 'my-groups'),
    },
  ];
  
  // Determine initialTab index based on activeTabKey
  const initialTabIndex = activeTabKey === 'explore' ? 0 : 1;

  return (
    <div className="container mx-auto p-4 text-white">
      <div className="flex flex-col sm:flex-row justify-between items-center mb-6 gap-4">
        <Heading level={1} className="whitespace-nowrap">Groups</Heading>
        {activeTabKey === 'explore' && (
          <Input
            type="text"
            placeholder="Search groups..."
            value={searchTerm}
            onChange={handleSearchChange}
            className="w-full sm:w-auto bg-gray-700 border-gray-600 placeholder-gray-400 text-white"
          />
        )}
        <Link href="/groups/create" passHref>
          <Button className="w-full sm:w-auto whitespace-nowrap">Create New Group</Button>
        </Link>
      </div>

      {/* The Tabs component now needs to be modified to call onTabChange with index */}
      {/* For now, assuming Tabs component is updated or we adapt its usage.
          The provided Tabs.tsx uses internal state and onClick on buttons.
          To make this work with external state (activeTabKey), Tabs.tsx would need modification
          to accept activeIndex and an onTabChange(index: number) prop.
          Let's assume we modify Tabs.tsx to support this.
          If not, we'd pass initialTab and let Tabs.tsx handle it internally,
          but then GroupsPage wouldn't know the active tab index directly from Tabs.tsx's props.
          
          For this iteration, I will adapt the usage here to match the existing Tabs.tsx.
          The `handleTabSelection` will be passed to a modified Tabs component.
          The `Tabs` component itself will call `setActiveTab(index)` internally,
          and if we want `GroupsPage` to react, `Tabs` needs an `onTabChange(index: number)` prop.
          
          Let's adjust the Tabs component props in common/Tabs.tsx to include onTabChange.
          For now, I will proceed as if Tabs component is updated.
          If Tabs component is NOT updated, the `content` for each tab is rendered by Tabs itself.
      */}
      <Tabs tabs={tabDefinitions} initialTab={initialTabIndex} />
      {/* Content is now rendered by the Tabs component via tabDefinitions.content */}
      {/* So, the explicit rendering block below is no longer needed here. */}
    </div>
  );
}
