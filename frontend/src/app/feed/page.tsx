'use client';
import { useState, useEffect, useCallback } from 'react';
import PostCard from '@/components/common/PostCard'; // Updated import path
import { Post } from '@/types/Post';
import { useRequest } from '@/hooks/useRequest';
import { Button } from '@/components/ui/button'; 

const FEED_LIMIT = 10; 

export default function FeedPage() {
  const [activeTab, setActiveTab] = useState<'forYou' | 'explore'>('explore');
  const [explorePosts, setExplorePosts] = useState<Post[]>([]);
  const [exploreOffset, setExploreOffset] = useState(0);
  const [canLoadMoreExplore, setCanLoadMoreExplore] = useState(true);
  const [isLoadingExplore, setIsLoadingExplore] = useState(false);
  const [errorExplore, setErrorExplore] = useState<string | null>(null);

  const { get } = useRequest<{ posts: Post[], count: number, limit: number, offset: number }>(); // Type argument moved here

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

  useEffect(() => {
    if (activeTab === 'explore') {
      fetchExplorePosts(0, true); // Initial load for explore tab
    } else {
      // Optionally clear explore posts or handle other tab logic
      setExplorePosts([]);
      setExploreOffset(0);
      setCanLoadMoreExplore(true);
    }
  }, [activeTab, fetchExplorePosts]); // fetchExplorePosts is stable due to useCallback
  
  const handleLoadMoreExplore = () => {
    if (canLoadMoreExplore && !isLoadingExplore) {
      fetchExplorePosts(exploreOffset);
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
        <div className="text-center text-gray-400 py-4">
          <p>"For You" feed coming soon!</p>
        </div>
      )}
    </div>
  );
}
