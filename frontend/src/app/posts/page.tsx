'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRequest } from '../../hooks/useRequest';
import { Post } from '../../types/Post';
import PostCard from '../../components/common/PostCard';
import { Button } from '../../components/ui/button';
import CreatePostForm from '../../components/profile/CreatePostForm';

interface PostsApiResponse {
  posts: Post[];
  limit: number;
  offset: number;
  total_posts: number; // This can be used for a more precise hasMore, but feed uses length === limit
}

const POSTS_PER_PAGE = 10;

export default function PostsPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoadingInitial, setIsLoadingInitial] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [nextPageOffset, setNextPageOffset] = useState(0);
  const [hasMore, setHasMore] = useState(true);

  const { get: fetchPostsRequest, error: fetchHookError } = useRequest<PostsApiResponse>();

  const loadPosts = useCallback(async (currentOffset: number) => {
    if (currentOffset === 0) {
      setIsLoadingInitial(true);
    } else {
      setIsFetchingMore(true);
    }
    setError(null); // Clear previous errors for this specific load attempt

    try {
      const data = await fetchPostsRequest(`/api/posts?limit=${POSTS_PER_PAGE}&offset=${currentOffset}`);
      if (data && data.posts) {
        setPosts(prevPosts => currentOffset === 0 ? data.posts : [...prevPosts, ...data.posts]);
        setNextPageOffset(currentOffset + data.posts.length); // Next offset is current + number of posts fetched
        setHasMore(data.posts.length === POSTS_PER_PAGE); // If fewer posts than limit are returned, no more posts
      } else {
        // If data or data.posts is null/undefined, assume no more posts for this request
        setHasMore(false);
        if (fetchHookError) { // If useRequest hook had an error
            setError(fetchHookError);
        }
      }
    } catch (err: any) {
      setError(err);
      setHasMore(false); // Stop trying to load more on a caught error
    } finally {
      if (currentOffset === 0) {
        setIsLoadingInitial(false);
      } else {
        setIsFetchingMore(false);
      }
    }
  }, [fetchPostsRequest, fetchHookError]);

  useEffect(() => {
    loadPosts(0); // Initial load
  }, [loadPosts]); // loadPosts will re-run if fetchPostsRequest or fetchHookError changes

  // Effect to handle errors from the useRequest hook if they occur outside a loadPosts call
  useEffect(() => {
    if (fetchHookError) {
      setError(fetchHookError);
      // Reset loading states if an error occurs
      if (nextPageOffset === 0) setIsLoadingInitial(false);
      else setIsFetchingMore(false);
      setHasMore(false); // Assume no more can be loaded if there's a persistent hook error
    }
  }, [fetchHookError, nextPageOffset]);

  const handleLoadMore = () => {
    if (hasMore && !isFetchingMore && !isLoadingInitial) {
      loadPosts(nextPageOffset);
    }
  };

  // Initial loading state: shown only when no posts are loaded yet.
  if (isLoadingInitial && posts.length === 0) {
    return <div className="text-center py-10">Loading posts...</div>;
  }

  // Error state: shown prominently if it's an initial load error and no posts are loaded.
  if (error && posts.length === 0 && !isLoadingInitial) { // ensure not still in initial loading phase
    return <div className="text-center py-10 text-red-500">Error loading posts: {error.message}</div>;
  }

  return (
    <div className="max-w-2xl mx-auto py-8 px-4 text-white">
      <h1 className="text-3xl font-bold text-white mb-6 text-center">All Posts</h1>
      <div className="mb-8">
        <CreatePostForm />
      </div>
      
      {/* No posts message: Show if not loading, no error, no posts, and no more to fetch */}
      {posts.length === 0 && !isLoadingInitial && !isFetchingMore && !error && !hasMore && (
         <div className="text-center py-10">No posts found.</div>
      )}

      <div className="space-y-6">
        {posts.map(post => (
          <PostCard key={post.id} post={post} />
        ))}
      </div>

      {/* Loading more indicator */}
      {isFetchingMore && <p className="text-center text-gray-400 py-4">Loading more posts...</p>}
      
      {/* Error message if some posts are loaded but loading more failed */}
      {error && !isFetchingMore && posts.length > 0 && (
         <p className="text-center text-red-500 py-4">Error loading more posts: {error.message}</p>
      )}

      {/* Load More Button: Show if hasMore, not fetching, and not initial loading */}
      {hasMore && !isFetchingMore && !isLoadingInitial && (
        <div className="text-center mt-8">
          <Button onClick={handleLoadMore} disabled={isFetchingMore || isLoadingInitial}>
            {isFetchingMore ? 'Loading...' : 'Load More'}
          </Button>
        </div>
      )}
      {/* No More Posts Message: Show if no more to load, not fetching, not initial loading, and some posts exist */}
      {!hasMore && !isFetchingMore && !isLoadingInitial && posts.length > 0 && (
        <p className="text-center text-gray-500 py-4 mt-4">No more posts to load.</p>
      )}
    </div>
  );
}
