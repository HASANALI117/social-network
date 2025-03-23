"use client";

import { useState } from 'react';
import { Post } from '@/types/Post';
import PostList from '@/components/profile/PostList';
import { FiCompass, FiZap } from 'react-icons/fi';

interface FeedTabsProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

function FeedTabs({ activeTab, onTabChange }: FeedTabsProps) {
  return (
    <div className="flex border-b border-gray-700 mb-6">
      <button
        className={`px-6 py-3 flex items-center gap-2 ${
          activeTab === 'for-you'
            ? 'border-b-2 border-purple-500 text-purple-400'
            : 'text-gray-400 hover:text-purple-400'
        }`}
        onClick={() => onTabChange('for-you')}
      >
        <FiZap />
        For You
      </button>
      <button
        className={`px-6 py-3 flex items-center gap-2 ${
          activeTab === 'explore'
            ? 'border-b-2 border-purple-500 text-purple-400'
            : 'text-gray-400 hover:text-purple-400'
        }`}
        onClick={() => onTabChange('explore')}
      >
        <FiCompass />
        Explore
      </button>
    </div>
  );
}

const dummyPosts: Post[] = Array.from({ length: 10 }, (_, i) => ({
  id: `dummy-${i + 1}`,
  user_id: 'user-1',
  title: `Post ${i + 1}`,
  content: `This is a dummy post ${i + 1} with some interesting content. Lorem ipsum dolor sit amet, consectetur adipiscing elit.`,
  image_url: i % 3 === 0 ? `https://picsum.photos/seed/${i + 1}/800/400` : undefined,
  privacy: i % 2 === 0 ? 'public' : 'friends',
  createdAt: new Date(Date.now() - i * 24 * 60 * 60 * 1000) // Posts from last 10 days
}));

const exploreData = [...dummyPosts].sort(() => Math.random() - 0.5);
const forYouData = [...dummyPosts].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());

export default function FeedPage() {
  const [activeTab, setActiveTab] = useState('for-you');
  const [currentPage, setCurrentPage] = useState(1);
  const postsPerPage = 5;

  const posts = activeTab === 'for-you' ? forYouData : exploreData;
  const totalPages = Math.ceil(posts.length / postsPerPage);
  const currentPosts = posts.slice(
    (currentPage - 1) * postsPerPage,
    currentPage * postsPerPage
  );

  return (
    <div className="max-w-4xl mx-auto p-6 min-h-screen bg-gray-900 text-white">
      <div className="mb-8">
        <h1 className="text-2xl font-bold mb-6">Feed</h1>
        <FeedTabs activeTab={activeTab} onTabChange={setActiveTab} />
      </div>

      <PostList posts={currentPosts} />

      {/* Pagination */}
      <div className="mt-8 flex justify-center gap-2">
        <button
          onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
          disabled={currentPage === 1}
          className="px-4 py-2 bg-gray-800 rounded-lg hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Previous
        </button>
        <div className="flex gap-2">
          {Array.from({ length: totalPages }, (_, i) => (
            <button
              key={i + 1}
              onClick={() => setCurrentPage(i + 1)}
              className={`w-10 h-10 rounded-lg ${
                currentPage === i + 1
                  ? 'bg-purple-600 text-white'
                  : 'bg-gray-800 hover:bg-gray-700'
              }`}
            >
              {i + 1}
            </button>
          ))}
        </div>
        <button
          onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
          disabled={currentPage === totalPages}
          className="px-4 py-2 bg-gray-800 rounded-lg hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Next
        </button>
      </div>
    </div>
  );
}
