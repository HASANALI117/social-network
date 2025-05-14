'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { useRequest } from '../../../../hooks/useRequest';
import { Post } from '../../../../types/Post';
import { Group } from '../../../../types/Group';
import PostCard from '../../../../components/common/PostCard'; // Though PostList will use this
import PostList from '../../../../components/common/PostList';
import { Button } from '../../../../components/ui/button';
import CreatePostForm from '../../../../components/profile/CreatePostForm';
import { Heading } from '../../../../components/ui/heading';

interface PostsApiResponse {
  posts: Post[];
  limit: number;
  offset: number;
  total_posts: number;
}

const POSTS_PER_PAGE = 10;

export default function GroupPostsPage() {
  const params = useParams();
  const groupId = params.id as string;

  const [posts, setPosts] = useState<Post[]>([]);
  const [group, setGroup] = useState<Group | null>(null);
  const [isLoadingInitialPosts, setIsLoadingInitialPosts] = useState(true);
  const [isLoadingGroupDetails, setIsLoadingGroupDetails] = useState(true);
  const [isFetchingMorePosts, setIsFetchingMorePosts] = useState(false);
  const [postsError, setPostsError] = useState<Error | null>(null);
  const [groupError, setGroupError] = useState<Error | null>(null);
  const [nextPageOffset, setNextPageOffset] = useState(0);
  const [hasMorePosts, setHasMorePosts] = useState(true);

  const { get: fetchPostsRequest, error: fetchPostsHookError } = useRequest<PostsApiResponse>();
  const { get: fetchGroupDetailsRequest, error: fetchGroupHookError } = useRequest<Group>();

  useEffect(() => {
    if (groupId) {
      const loadGroupDetails = async () => {
        setIsLoadingGroupDetails(true);
        setGroupError(null);
        try {
          const groupData = await fetchGroupDetailsRequest(`/api/groups/${groupId}`);
          if (groupData) {
            setGroup(groupData);
          } else if (fetchGroupHookError) {
            setGroupError(fetchGroupHookError);
          } else {
            setGroupError(new Error('Group not found or failed to load.'));
          }
        } catch (err: any) {
          setGroupError(err);
        } finally {
          setIsLoadingGroupDetails(false);
        }
      };
      loadGroupDetails();
    }
  }, [groupId, fetchGroupDetailsRequest, fetchGroupHookError]);

  const loadPosts = useCallback(async (currentOffset: number) => {
    if (!groupId) return;

    if (currentOffset === 0) {
      setIsLoadingInitialPosts(true);
    } else {
      setIsFetchingMorePosts(true);
    }
    setPostsError(null);

    try {
      const data = await fetchPostsRequest(`/api/groups/${groupId}/posts?limit=${POSTS_PER_PAGE}&offset=${currentOffset}`);
      if (data && data.posts) {
        setPosts(prevPosts => currentOffset === 0 ? data.posts : [...prevPosts, ...data.posts]);
        setNextPageOffset(currentOffset + data.posts.length);
        setHasMorePosts(data.posts.length === POSTS_PER_PAGE);
      } else {
        setHasMorePosts(false);
        if (fetchPostsHookError) {
            setPostsError(fetchPostsHookError);
        }
      }
    } catch (err: any) {
      setPostsError(err);
      setHasMorePosts(false);
    } finally {
      if (currentOffset === 0) {
        setIsLoadingInitialPosts(false);
      } else {
        setIsFetchingMorePosts(false);
      }
    }
  }, [groupId, fetchPostsRequest, fetchPostsHookError]);

  useEffect(() => {
    if (groupId) {
      loadPosts(0); 
    }
  }, [groupId, loadPosts]);

  useEffect(() => {
    if (fetchPostsHookError) {
      setPostsError(fetchPostsHookError);
      if (nextPageOffset === 0) setIsLoadingInitialPosts(false);
      else setIsFetchingMorePosts(false);
      setHasMorePosts(false);
    }
  }, [fetchPostsHookError, nextPageOffset]);

  const handleLoadMorePosts = () => {
    if (hasMorePosts && !isFetchingMorePosts && !isLoadingInitialPosts && groupId) {
      loadPosts(nextPageOffset);
    }
  };

  const handlePostCreated = useCallback((newPost: Post) => {
    setPosts(prevPosts => [newPost, ...prevPosts]);
    // Consider if total_posts needs adjustment for pagination if it's used for hasMore
  }, []);

  if (!groupId) {
    return <div className="text-center py-10 text-red-500">Group ID is missing from URL.</div>;
  }

  if (isLoadingGroupDetails) {
    return <div className="text-center py-10">Loading group details...</div>;
  }

  if (groupError && !group) {
      return <div className="text-center py-10 text-red-500">Error loading group details: {groupError.message}</div>;
  }
  
  const pageTitle = group ? `Posts for ${group.name}` : 'Group Posts';

  return (
    <div className="max-w-2xl mx-auto py-8 px-4 text-white">
      <Heading level={1} className="mb-6 text-center">{pageTitle}</Heading>
      
      <div className="mb-8">
        <CreatePostForm onSubmit={handlePostCreated} groupId={groupId} />
      </div>
      
      {isLoadingInitialPosts && posts.length === 0 && (
        <div className="text-center py-10">Loading posts...</div>
      )}

      {postsError && posts.length === 0 && !isLoadingInitialPosts && (
        <div className="text-center py-10 text-red-500">Error loading posts: {postsError.message}</div>
      )}

      {!isLoadingInitialPosts && !postsError && posts.length === 0 && !hasMorePosts && (
         <div className="text-center py-10">No posts found in this group yet. Be the first to post!</div>
      )}

      <PostList posts={posts} />

      {isFetchingMorePosts && <p className="text-center text-gray-400 py-4">Loading more posts...</p>}
      
      {postsError && !isFetchingMorePosts && posts.length > 0 && (
         <p className="text-center text-red-500 py-4">Error loading more posts: {postsError.message}</p>
      )}

      {hasMorePosts && !isFetchingMorePosts && !isLoadingInitialPosts && (
        <div className="text-center mt-8">
          <Button onClick={handleLoadMorePosts} disabled={isFetchingMorePosts || isLoadingInitialPosts}>
            {isFetchingMorePosts ? 'Loading...' : 'Load More'}
          </Button>
        </div>
      )}
      {!hasMorePosts && !isFetchingMorePosts && !isLoadingInitialPosts && posts.length > 0 && (
        <p className="text-center text-gray-500 py-4 mt-4">No more posts to load.</p>
      )}
    </div>
  );
}