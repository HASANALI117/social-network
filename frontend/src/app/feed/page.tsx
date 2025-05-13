'use client';
import { useState, useEffect, useCallback } from 'react';
import PostCard from '@/components/common/PostCard'; // Updated import path
import { Post } from '@/types/Post';
import { useRequest } from '@/hooks/useRequest';
import { Button } from '@/components/ui/button';
import { useUserStore } from '@/store/useUserStore';

const FEED_LIMIT = 10;

export default function FeedPage() {
  const [activeTab, setActiveTab] = useState<'forYou' | 'explore'>('explore');
  const [explorePosts, setExplorePosts] = useState<Post[]>([]);
  const [exploreOffset, setExploreOffset] = useState(0);
  const [canLoadMoreExplore, setCanLoadMoreExplore] = useState(true);
  const [isLoadingExplore, setIsLoadingExplore] = useState(false);
  const [errorExplore, setErrorExplore] = useState<string | null>(null);

  // State for "For You" tab
  const [forYouPosts, setForYouPosts] = useState<Post[]>([]);
  const [forYouOffset, setForYouOffset] = useState(0);
  const [canLoadMoreForYou, setCanLoadMoreForYou] = useState(true);
  const [isLoadingForYou, setIsLoadingForYou] = useState(false);
  const [errorForYou, setErrorForYou] = useState<string | null>(null);

  const { user, isAuthenticated, hydrated } = useUserStore();

  const { get } = useRequest<{ posts: Post[], count: number, limit: number, offset: number }>();

  const fetchExplorePosts = useCallback(async (currentOffset: number, isInitialLoad: boolean = false) => {
    setIsLoadingExplore(true);
    setErrorExplore(null);
    try {
      const response = await get( // Removed type argument from here
        `/api/posts/explore?limit=${FEED_LIMIT}&offset=${currentOffset}`
      );
      if (response && response.posts) {
        setExplorePosts(prev => isInitialLoad ? response.posts : [...prev, ...response.posts]);
        setExploreOffset(currentOffset + response.posts.length);
        setCanLoadMoreExplore(response.posts.length === FEED_LIMIT);
      } else {
        setCanLoadMoreExplore(false);
      }
    } catch (err: any) {
      setErrorExplore(err.message || 'Failed to fetch explore posts');
      setCanLoadMoreExplore(false); // Stop loading on error
    } finally {
      setIsLoadingExplore(false);
    }
  }, [get]);

  const fetchForYouPosts = useCallback(async (currentOffset: number, isInitialLoad: boolean = false) => {
    if (!isAuthenticated || !user) { // Don't fetch if not logged in or user object is null
      setForYouPosts([]);
      setCanLoadMoreForYou(false);
      setErrorForYou(null); // Clear any previous error
      return;
    }
    setIsLoadingForYou(true);
    setErrorForYou(null);
    try {
      // Type argument for `get` is defined when `useRequest` is called, not here.
      const response = await get(
        `/api/posts/following?limit=${FEED_LIMIT}&offset=${currentOffset}`
      );
      if (response && response.posts) {
        setForYouPosts(prev => isInitialLoad ? response.posts : [...prev, ...response.posts]);
        setForYouOffset(currentOffset + response.posts.length);
        setCanLoadMoreForYou(response.posts.length === FEED_LIMIT);
      } else {
        setCanLoadMoreForYou(false);
      }
    } catch (err: any) {
      setErrorForYou(err.message || 'Failed to fetch your feed');
      setCanLoadMoreForYou(false);
    } finally {
      setIsLoadingForYou(false);
    }
  }, [get, user]); // Add user to dependencies

  useEffect(() => {
    if (activeTab === 'explore') {
      fetchExplorePosts(0, true); // Initial load for explore tab
      // Clear "For You" posts when switching to explore
      setForYouPosts([]);
      setForYouOffset(0);
      setCanLoadMoreForYou(true);
      setErrorForYou(null);
    } else if (activeTab === 'forYou') {
      if (hydrated) { // Only fetch if auth state is resolved (store is hydrated)
        if (isAuthenticated && user) {
          fetchForYouPosts(0, true); // Initial load for "For You" tab
        } else {
          setForYouPosts([]); // Clear posts if user logs out or is not authenticated
          setForYouOffset(0);
          setCanLoadMoreForYou(false); // Cannot load more if not logged in
          setErrorForYou(null); // Clear any previous errors
        }
      }
      // Clear explore posts when switching to "For You"
      setExplorePosts([]);
      setExploreOffset(0);
      setCanLoadMoreExplore(true);
      setErrorExplore(null);
    }
  }, [activeTab, fetchExplorePosts, fetchForYouPosts, user, isAuthenticated, hydrated]);
  
  const handleLoadMoreExplore = () => {
    if (canLoadMoreExplore && !isLoadingExplore) {
      fetchExplorePosts(exploreOffset);
    }
  };

  const handleLoadMoreForYou = () => {
    if (canLoadMoreForYou && !isLoadingForYou && isAuthenticated && user) {
      fetchForYouPosts(forYouOffset);
    }
  };

  return (
    <div className="max-w-2xl mx-auto py-8 px-4 text-white"> {/* Added text-white for better visibility on dark bg */}
      <div className="mb-6 flex border-b border-gray-700">
        <button
          className={`py-2 px-4 text-lg font-medium ${activeTab === 'forYou' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-400 hover:text-gray-200'}`}
          onClick={() => setActiveTab('forYou')}
        >
          For You
        </button>
        <button
          className={`py-2 px-4 text-lg font-medium ${activeTab === 'explore' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-400 hover:text-gray-200'}`}
          onClick={() => setActiveTab('explore')}
        >
          Explore
        </button>
      </div>

      {activeTab === 'explore' && (
        <div>
          {explorePosts.map(post => (
            <PostCard key={post.id} post={post} />
          ))}
          {isLoadingExplore && <p className="text-center text-gray-400 py-4">Loading posts...</p>}
          {errorExplore && <p className="text-center text-red-500 py-4">{errorExplore}</p>}
          {!isLoadingExplore && explorePosts.length === 0 && !errorExplore && (
            <p className="text-center text-gray-400 py-4">No posts to explore yet.</p>
          )}
          {canLoadMoreExplore && !isLoadingExplore && explorePosts.length > 0 && (
            <div className="text-center mt-6">
              <Button onClick={handleLoadMoreExplore} outline disabled={isLoadingExplore}>
                {isLoadingExplore ? 'Loading...' : 'Load More'}
              </Button>
            </div>
          )}
           {!canLoadMoreExplore && !isLoadingExplore && explorePosts.length > 0 && (
            <p className="text-center text-gray-500 py-4">No more posts to load.</p>
          )}
        </div>
      )}

      {activeTab === 'forYou' && (
        <div>
          {!hydrated && <p className="text-center text-gray-400 py-4">Loading user session...</p>}
          {hydrated && !isAuthenticated && (
            <p className="text-center text-gray-400 py-4">
              Please <a href="/login" className="text-blue-500 hover:underline">log in</a> to see posts from users you follow.
            </p>
          )}
          {hydrated && isAuthenticated && user && (
            <>
              {forYouPosts.map(post => (
                <PostCard key={post.id} post={post} />
              ))}
              {isLoadingForYou && <p className="text-center text-gray-400 py-4">Loading your feed...</p>}
              {errorForYou && <p className="text-center text-red-500 py-4">{errorForYou}</p>}
              {!isLoadingForYou && forYouPosts.length === 0 && !errorForYou && (
                <p className="text-center text-gray-400 py-4">
                  No posts from users you follow yet. Explore and follow some interesting people!
                </p>
              )}
              {canLoadMoreForYou && !isLoadingForYou && forYouPosts.length > 0 && (
                <div className="text-center mt-6">
                  <Button onClick={handleLoadMoreForYou} outline disabled={isLoadingForYou}>
                    {isLoadingForYou ? 'Loading...' : 'Load More'}
                  </Button>
                </div>
              )}
              {!canLoadMoreForYou && !isLoadingForYou && forYouPosts.length > 0 && (
                <p className="text-center text-gray-500 py-4">No more posts to load from your feed.</p>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
