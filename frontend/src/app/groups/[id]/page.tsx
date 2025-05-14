'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
// import { useForm, SubmitHandler } from 'react-hook-form'; // Removed for Phase 1
import { useRequest } from '../../../hooks/useRequest';
import { Group, PostSummary, EventSummary } from '../../../types/Group';
import { User, UserBasicInfo } from '../../../types/User';
// import { Post } from '../../../types/Post'; // Removed for Phase 1
// import { GroupInvitation } from '../../../types/GroupInvitation'; // Removed for Phase 1
// import { GroupJoinRequest } from '../../../types/GroupJoinRequest'; // Removed for Phase 1
import { useUserStore } from '../../../store/useUserStore';
import Tabs from '../../../components/common/Tabs'; // Added for Phase 2
import { Heading } from '../../../components/ui/heading';
import { Text } from '../../../components/ui/text';
// import { Textarea } from '../../../components/ui/textarea'; // Removed for Phase 1
import { Avatar } from '../../../components/ui/avatar';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input'; // Added for Phase 3b
import { Alert, AlertTitle, AlertDescription, AlertBody, AlertActions } from '../../../components/ui/alert'; // AlertDescription, AlertTitle removed as Alert is used simply
import { format } from 'date-fns'; // Added for Phase 2 to format dates

export default function GroupDetailPage() {
  const params = useParams();
  const router = useRouter();
  const groupId = params.id as string;

  const [group, setGroup] = useState<Group | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isMember, setIsMember] = useState<boolean>(false);

  // State for User Search and Invite UI (Phase 3b)
  const [showInviteUI, setShowInviteUI] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResults, setSearchResults] = useState<UserBasicInfo[]>([]);
  const [isActualSearchLoading, setIsActualSearchLoading] = useState(false); // To manage loading state for debounced search
  const [searchError, setSearchError] = useState<string | null>(null);
  
  const [invitingUserId, setInvitingUserId] = useState<string | null>(null);
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null);

  const { user: currentUser } = useUserStore(); // Kept for potential use in button actions
  const { get: fetchGroupRequest, error: groupApiError } = useRequest<Group>();
  const { get: searchUsersRequestHook, error: searchApiHookError, isLoading: isSearchHookLoading } = useRequest<UserBasicInfo[]>();
  const { post: sendInviteRequestHook, error: inviteApiHookError, isLoading: isInviteHookLoading } = useRequest<{ message: string }>();


  const loadGroupDetails = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);
    setGroup(null); // Reset group state before fetching
    setIsMember(false); // Reset member status
    try {
      const groupData = await fetchGroupRequest(`/api/groups/${id}`);
      if (groupData) {
        const processedGroupData = {
          ...groupData,
          posts: groupData.posts || [],
          members: groupData.members || [],
          events: groupData.events || [],
        };
        console.log("Processed group data before setGroup:", JSON.stringify(processedGroupData));
        setGroup(processedGroupData);

        // Determine isMember based on the original groupData from backend
        // to reflect if member-specific data was sent.
        if (groupData.members !== undefined && groupData.members !== null) {
          setIsMember(true);
        } else {
          setIsMember(false);
        }
      } else if (groupApiError) {
        setError(groupApiError.message || 'Failed to load group details.');
      } else {
        setError('Group not found or failed to load.');
      }
    } catch (err: any) {
      setError(err.message || 'An unexpected error occurred.');
    } finally {
      setIsLoading(false);
    }
  }, [fetchGroupRequest, groupApiError]); // Removed userApiError and fetchUserRequest as creator info is in group

  useEffect(() => {
    if (groupId) {
      loadGroupDetails(groupId);
    } else {
      setError("Group ID is missing.");
      setIsLoading(false);
    }
  }, [groupId, loadGroupDetails]);
  
  // This useEffect handles potential errors from the useRequest hook for fetching group details
  useEffect(() => {
    if (groupApiError && !isLoading && !group) { // Check !isLoading to avoid setting error during initial load
        setError(groupApiError.message || 'Failed to load group details.');
    }
  }, [groupApiError, isLoading, group]);


  // Placeholder for "Request to Join" button action
  const handleRequestToJoin = () => {
    // Functionality to be added in a later phase
    console.log('Request to Join button clicked');
    alert('Request to Join functionality will be implemented later.');
  };

  // Debounce utility function (Phase 3b)
  const debounce = <F extends (...args: any[]) => any>(func: F, waitFor: number) => {
    let timeout: ReturnType<typeof setTimeout> | null = null;
    const debounced = (...args: Parameters<F>) => {
      if (timeout !== null) {
        clearTimeout(timeout);
        timeout = null;
      }
      timeout = setTimeout(() => func(...args), waitFor);
    };
    return debounced as (...args: Parameters<F>) => void; // Ensure void return for React event handlers
  };

  // User Search and Invite functions (Phase 3b)
  const toggleInviteUI = () => {
    setShowInviteUI(prev => !prev);
    if (showInviteUI) { // Reset states when closing
      setSearchTerm('');
      setSearchResults([]);
      setSearchError(null);
      setInviteError(null);
      setInviteSuccess(null);
    }
  };

  const handleSearchUsers = useCallback(async (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      setSearchError(null);
      setIsActualSearchLoading(false);
      return;
    }
    setIsActualSearchLoading(true);
    setSearchError(null);
    setInviteError(null);
    setInviteSuccess(null);
    try {
      const results = await searchUsersRequestHook(`/api/users/search?q=${encodeURIComponent(query)}`);
      if (results) {
        setSearchResults(results);
      } else if (searchApiHookError) {
        setSearchError(searchApiHookError.message || 'Failed to search users.');
        setSearchResults([]);
      } else {
        setSearchResults([]);
      }
    } catch (err: any) {
      setSearchError(err.message || 'An unexpected error occurred during search.');
      setSearchResults([]);
    } finally {
      setIsActualSearchLoading(false);
    }
  }, [searchUsersRequestHook, searchApiHookError]);

  const debouncedSearchUsers = useCallback(debounce(handleSearchUsers, 400), [handleSearchUsers]);

  useEffect(() => {
    if (searchTerm.trim()) {
      debouncedSearchUsers(searchTerm);
    } else {
      setSearchResults([]);
      setSearchError(null);
      // If search term is cleared, stop any ongoing debounced call visual loading indicator
      if(isActualSearchLoading && !searchTerm.trim()){
          setIsActualSearchLoading(false);
      }
    }
  }, [searchTerm, debouncedSearchUsers, isActualSearchLoading]);

  const handleSendInvite = async (userIdToInvite: string) => {
    if (!groupId) return;
    setInvitingUserId(userIdToInvite);
    setInviteError(null);
    setInviteSuccess(null);
    try {
      const response = await sendInviteRequestHook(`/api/groups/${groupId}/invite`, { user_id: userIdToInvite });
      if (response) {
        const invitedUser = searchResults.find(u => u.user_id === userIdToInvite);
        setInviteSuccess(`Invitation sent to ${invitedUser ? `${invitedUser.first_name} ${invitedUser.last_name} (@${invitedUser.username})` : 'user'}!`);
        // Consider removing user from search results or disabling button permanently after successful invite
      } else if (inviteApiHookError) {
        setInviteError(inviteApiHookError.message || 'Failed to send invitation.');
      }
    } catch (err: any) {
      setInviteError(err.message || 'An unexpected error occurred while sending invitation.');
    } finally {
      setInvitingUserId(null);
    }
  };

  // Placeholder for "Leave Group" button action
  const handleLeaveGroup = () => {
    console.log('Leave Group button clicked');
    alert('Leave Group functionality will be implemented later.');
  };

  if (isLoading) {
    return <div className="container mx-auto p-4 text-center text-white"><Text>Loading group details...</Text></div>;
  }

  if (error) {
    return <div className="container mx-auto p-4 text-center text-red-500"><Text>Error: {error}</Text></div>;
  }

  if (!group) {
    return <div className="container mx-auto p-4 text-center text-white"><Text>Group not found.</Text></div>;
  }

  const {
    name,
    description,
    avatar_url,
    creator_info,
    created_at,
    members_count,
    posts_count,
    events_count,
    posts, // Added for Phase 2
    members, // Added for Phase 2
    events, // Added for Phase 2
  } = group;

  const creatorFullName = `${creator_info.first_name} ${creator_info.last_name}`;
  const creatorUsername = creator_info.username;
  const creatorAvatarUrl = creator_info.avatar_url;

  return (
    <div className="min-h-screen bg-gray-900 text-white p-4 md:p-8">
      <div className="container mx-auto max-w-4xl">
        {/* Group Header Component */}
        <header className="mb-8 p-6 bg-gray-800 rounded-lg shadow-xl">
          <div className="flex flex-col sm:flex-row items-center">
            <Avatar
              src={avatar_url || null}
              initials={!avatar_url && name ? name.substring(0, 1).toUpperCase() : undefined}
              alt={`${name} avatar`}
              className="h-24 w-24 md:h-32 md:w-32 rounded-full mr-0 sm:mr-6 mb-4 sm:mb-0 border-2 border-purple-500"
            />
            <div className="text-center sm:text-left">
              <Heading level={1} className="text-3xl md:text-4xl font-bold text-purple-400 mb-2">
                {name}
              </Heading>
              <Text className="text-gray-300 mb-3 text-lg">{description}</Text>
              <div className="flex items-center justify-center sm:justify-start text-sm text-gray-400">
                <Avatar
                  src={creatorAvatarUrl || null}
                  initials={!creatorAvatarUrl && creatorFullName ? creatorFullName.substring(0, 1).toUpperCase() : undefined}
                  alt={creatorFullName}
                  className="h-8 w-8 mr-2 rounded-full border border-gray-600"
                />
                <Text>
                  Created by:{' '}
                  <Link href={`/profile/${creator_info.user_id}`} className="text-purple-300 hover:underline">
                    {creatorFullName} ({creatorUsername})
                  </Link>
                </Text>
                <span className="mx-2 text-gray-500">|</span>
                <Text>Created on: {(() => {
                  console.log("Raw group created_at:", created_at);
                  if (created_at) {
                    try {
                      const dateObj = new Date(created_at);
                      if (!isNaN(dateObj.getTime())) {
                        return dateObj.toLocaleDateString();
                      } else {
                        console.error("Invalid date string for group created_at:", created_at);
                        return 'Invalid date';
                      }
                    } catch (e) {
                      console.error("Error parsing date string for group created_at:", created_at, e);
                      return 'Error parsing date';
                    }
                  }
                  return 'Date not available';
                })()}</Text>
              </div>
            </div>
          </div>
        </header>

        {/* Conditional Content Area */}
        <section className="p-6 bg-gray-800 rounded-lg shadow-xl">
          {!isMember ? (
            // Non-Member View
            <div className="text-center">
              <Heading level={2} className="text-2xl mb-4 text-gray-200">
                Join the Conversation!
              </Heading>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6 text-lg">
                <div className="p-4 bg-gray-700 rounded-md">
                  <Text className="font-semibold text-purple-400">{members_count}</Text>
                  <Text className="text-gray-300">Members</Text>
                </div>
                <div className="p-4 bg-gray-700 rounded-md">
                  <Text className="font-semibold text-purple-400">{posts_count}</Text>
                  <Text className="text-gray-300">Posts</Text>
                </div>
                <div className="p-4 bg-gray-700 rounded-md">
                  <Text className="font-semibold text-purple-400">{events_count}</Text>
                  <Text className="text-gray-300">Events</Text>
                </div>
              </div>
              <Button
                onClick={handleRequestToJoin}
                className="bg-purple-600 hover:bg-purple-700 text-white font-bold py-3 px-6 rounded-lg text-lg"
              >
                Request to Join
              </Button>
            </div>
          ) : (
            // Member View
            <div>
              {/* Stats remain visible for members */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8 text-lg text-center">
                <div className="p-4 bg-gray-700 rounded-md">
                  <Text className="font-semibold text-purple-400">{members_count}</Text>
                  <Text className="text-gray-300">Members</Text>
                </div>
                <div className="p-4 bg-gray-700 rounded-md">
                  <Text className="font-semibold text-purple-400">{posts_count}</Text>
                  <Text className="text-gray-300">Posts</Text>
                </div>
                <div className="p-4 bg-gray-700 rounded-md">
                  <Text className="font-semibold text-purple-400">{events_count}</Text>
                  <Text className="text-gray-300">Events</Text>
                </div>
              </div>

              <Tabs tabs={[
                {
                  label: 'Posts',
                  content: (
                    <Tabs.Panel id="posts" className="py-4">
                      {(() => {
                        // Granular logging
                        console.log("Posts Tab: group is", group ? "defined" : "undefined");
                        if (group) {
                          console.log("Posts Tab: group.posts is", group.posts ? "defined array" : "undefined or not array", group.posts);
                          if (group.posts) {
                            console.log("Posts Tab: group.posts.length is", group.posts.length);
                          }
                        }

                        // Simplified conditional rendering logic
                        const postsArray = group?.posts; // Use optional chaining to get the array

                        if (postsArray && postsArray.length > 0) {
                          return (
                            <div className="space-y-4">
                              {postsArray.map((post: PostSummary) => {
                                console.log("Raw post created_at:", post.created_at);
                                let formattedPostDate = 'Date not available';
                                if (post.created_at) {
                                  try {
                                    const dateObj = new Date(post.created_at);
                                    if (!isNaN(dateObj.getTime())) {
                                      formattedPostDate = format(dateObj, 'PPpp');
                                    } else {
                                      console.error("Invalid date string for post:", post.created_at);
                                      formattedPostDate = 'Invalid date';
                                    }
                                  } catch (e) {
                                    console.error("Error parsing date string for post:", post.created_at, e);
                                    formattedPostDate = 'Error parsing date';
                                  }
                                }
                                return (
                                  <div key={post.id} className="p-4 bg-gray-700 rounded-lg shadow">
                                    <div className="flex items-center mb-2">
                                      <Avatar
                                        src={post.creator_avatar_url || null}
                                        initials={!post.creator_avatar_url && post.creator_name ? post.creator_name.substring(0,1).toUpperCase() : undefined}
                                        alt={post.creator_name}
                                        className="h-10 w-10 mr-3 rounded-full"
                                      />
                                      <div>
                                        <Text className="font-semibold text-purple-300">{post.creator_name}</Text>
                                        <Text className="text-xs text-gray-400">
                                          {formattedPostDate}
                                        </Text>
                                      </div>
                                    </div>
                                    <Text className="text-gray-300 whitespace-pre-wrap">
                                      {/* {post.content.length > 150 ? `${post.content.substring(0, 150)}...` : post.content} */}
                                    </Text>
                                    {post.image_url && (
                                        <div className="mt-2">
                                            <img src={post.image_url} alt="Post image" className="max-h-60 rounded-md object-cover" />
                                        </div>
                                    )}
                                     <Link href={`/posts/${post.id}`} className="text-sm text-purple-400 hover:underline mt-2 inline-block">
                                      View Post
                                    </Link>
                                  </div>
                                );
                              })}
                            </div>
                          );
                        } else {
                          // This will also catch if postsArray is undefined or null
                          return <Text className="text-center text-gray-400 py-4">No posts in this group yet.</Text>;
                        }
                      })()}
                    </Tabs.Panel>
                  )
                },
                {
                  label: 'Members',
                  content: (
                    <Tabs.Panel id="members" className="py-4">
                      {(() => {
                        // Granular logging
                        console.log("Members Tab: group is", group ? "defined" : "undefined");
                        if (group) {
                          console.log("Members Tab: group.members is", group.members ? "defined array" : "undefined or not array", group.members);
                          if (group.members) {
                            console.log("Members Tab: group.members.length is", group.members.length);
                          }
                        }

                        // Simplified conditional rendering logic
                        const membersArray = group?.members; // Use optional chaining to get the array

                        if (membersArray && membersArray.length > 0) {
                          return (
                            <div className="space-y-3">
                              {membersArray.map((member: UserBasicInfo) => (
                                <div key={member.user_id} className="flex items-center p-3 bg-gray-700 rounded-lg shadow">
                                  <Avatar
                                    src={member.avatar_url || null}
                                    initials={!member.avatar_url && member.first_name ? `${member.first_name.substring(0,1)}${member.last_name.substring(0,1)}`.toUpperCase() : undefined}
                                    alt={`${member.first_name} ${member.last_name}`}
                                    className="h-10 w-10 mr-3 rounded-full"
                                  />
                                  <div>
                                    <Link href={`/profile/${member.user_id}`} className="text-purple-300 hover:underline font-semibold">
                                      {member.first_name} {member.last_name}
                                    </Link>
                                    <Text className="text-xs text-gray-400">@{member.username}</Text>
                                  </div>
                                </div>
                              ))}
                            </div>
                          );
                        } else {
                          // This will also catch if membersArray is undefined or null
                          return <Text className="text-center text-gray-400 py-4">No members to display.</Text>;
                        }
                      })()}
                    </Tabs.Panel>
                  )
                },
                {
                  label: 'Events',
                  content: (
                    <Tabs.Panel id="events" className="py-4">
                      {(() => {
                        // Granular logging
                        console.log("Events Tab: group is", group ? "defined" : "undefined");
                        if (group) {
                          console.log("Events Tab: group.events is", group.events ? "defined array" : "undefined or not array", group.events);
                          if (group.events) {
                            console.log("Events Tab: group.events.length is", group.events.length);
                          }
                        }

                        // Simplified conditional rendering logic
                        const eventsArray = group?.events; // Use optional chaining to get the array

                        if (eventsArray && eventsArray.length > 0) {
                          return (
                            <div className="space-y-4">
                              {eventsArray.map((event, index) => (
                                // SIMPLIFIED RENDERING:
                                <div key={event.event_id || index}>
                                  <Text>Event Title: {event.title || "No title"}</Text>
                                  <Text>Event ID: {event.event_id}</Text>
                                  {/* Do NOT try to render event.event_time here for now unless fully safeguarded */}
                                  <Text>Raw Event Time (from log): {event.event_time === undefined ? "undefined" : event.event_time || "empty/null"}</Text>
                                </div>
                              ))}
                            </div>
                          );
                        } else {
                          return <Text>No events in this group yet.</Text>;
                        }
                      })()}
                    </Tabs.Panel>
                  )
                },
                {
                  label: 'Invitations',
                  content: (
                    <Text className="text-center text-gray-400 py-4">
                      Group Invitations and Join Requests management will go here.
                    </Text>
                  )
                }
              ]} />

              <div className="flex flex-col sm:flex-row justify-center space-y-3 sm:space-y-0 sm:space-x-4 mt-8">
                <Button
                  onClick={toggleInviteUI}
                  className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-md"
                >
                  {showInviteUI ? 'Close Invite UI' : 'Invite User'}
                </Button>
                <Button
                  onClick={handleLeaveGroup}
                  color="red"
                  className="bg-red-600 hover:bg-red-700 text-white font-semibold py-2 px-4 rounded-md"
                >
                  Leave Group
                </Button>
              </div>

              {showInviteUI && (
                <div className="mt-6 p-4 bg-gray-700 rounded-lg shadow-md">
                  <Heading level={3} className="text-xl mb-4 text-gray-100">Invite Users to Group</Heading>
                  <Input
                    type="text"
                    placeholder="Search users by name or username..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="mb-4 w-full bg-gray-800 border-gray-600 placeholder-gray-500 text-white focus:ring-purple-500 focus:border-purple-500"
                  />

                  {isActualSearchLoading && <Text className="text-sm text-gray-400 my-2">Searching...</Text>}
                  
                  {searchError && (
                    <Alert open={!!searchError} onClose={() => setSearchError(null)} size="sm">
                      <AlertTitle className="text-red-600">Search Error</AlertTitle>
                      <AlertDescription>{searchError}</AlertDescription>
                      <AlertActions>
                        <Button plain onClick={() => setSearchError(null)}>OK</Button>
                      </AlertActions>
                    </Alert>
                  )}
                  
                  {inviteSuccess && (
                    <Alert open={!!inviteSuccess} onClose={() => setInviteSuccess(null)} size="sm">
                      <AlertTitle className="text-green-600">Success</AlertTitle>
                      <AlertDescription>{inviteSuccess}</AlertDescription>
                      <AlertActions>
                        <Button plain onClick={() => setInviteSuccess(null)}>OK</Button>
                      </AlertActions>
                    </Alert>
                  )}

                  {inviteError && (
                    <Alert open={!!inviteError} onClose={() => setInviteError(null)} size="sm">
                      <AlertTitle className="text-red-600">Invite Error</AlertTitle>
                      <AlertDescription>{inviteError}</AlertDescription>
                      <AlertActions>
                        <Button plain onClick={() => setInviteError(null)}>OK</Button>
                      </AlertActions>
                    </Alert>
                  )}

                  {searchResults.length > 0 && !isActualSearchLoading && (
                    <div className="space-y-2 max-h-72 overflow-y-auto pr-2">
                      {searchResults.map(userResult => (
                        <div key={userResult.user_id} className="flex items-center justify-between p-3 bg-gray-600 rounded-md hover:bg-gray-500 transition-colors">
                          <div className="flex items-center">
                            <Avatar
                              src={userResult.avatar_url || null}
                              initials={!userResult.avatar_url && userResult.first_name && userResult.last_name ? `${userResult.first_name[0]}${userResult.last_name[0]}`.toUpperCase() : userResult.username ? userResult.username[0].toUpperCase() : 'U'}
                              alt={`${userResult.first_name} ${userResult.last_name}`}
                              className="h-10 w-10 mr-3 rounded-full border-2 border-gray-500"
                            />
                            <div>
                              <Text className="text-base font-semibold text-purple-300">{userResult.first_name} {userResult.last_name}</Text>
                              <Text className="text-xs text-gray-400">@{userResult.username}</Text>
                            </div>
                          </div>
                          <Button
                            onClick={() => handleSendInvite(userResult.user_id)}
                            disabled={(isInviteHookLoading && invitingUserId === userResult.user_id) || (!!inviteSuccess && invitingUserId === userResult.user_id) } // Disable if inviting or successfully invited
                            className={`text-white font-semibold py-1 px-3 rounded-md text-sm ${
                              (isInviteHookLoading && invitingUserId === userResult.user_id) ? 'bg-gray-500' :
                              (!!inviteSuccess && invitingUserId === userResult.user_id) ? 'bg-green-700 cursor-not-allowed' :
                              'bg-green-600 hover:bg-green-700'
                            }`}
                          >
                            {(isInviteHookLoading && invitingUserId === userResult.user_id) ? 'Inviting...' :
                             (!!inviteSuccess && invitingUserId === userResult.user_id && !inviteError) ? 'Invited!' : 'Invite'}
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}
                  {searchTerm && !isActualSearchLoading && searchResults.length === 0 && !searchError && (
                    <Text className="text-sm text-gray-400 my-2">No users found matching your search.</Text>
                  )}
                </div>
              )}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}