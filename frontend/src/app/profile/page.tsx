"use client"

import { useState } from 'react';
import { User } from '@/types/User';
import { FiEdit, FiUsers, FiLock, FiUnlock, FiPlus, FiMessageSquare, FiHeart, FiShare } from 'react-icons/fi';

const dummyUser: User = {
  id: '1',
  username: 'johndoe',
  email: 'john@example.com',
  first_name: 'John',
  last_name: 'Doe',
  avatar_url: null,
  about_me: 'Frontend developer passionate about creating beautiful user experiences',
  birth_date: '1990-01-01',
  created_at: '2024-01-01',
  updated_at: '2024-01-01'
};

export default function ProfilePage() {
  const [isPublic, setIsPublic] = useState(true);
  const [activeTab, setActiveTab] = useState('posts');
  const [posts, setPosts] = useState([
    {
      id: 1,
      content: 'Just launched my new portfolio website! 🚀',
      likes: 42,
      comments: 12,
      timestamp: '2024-03-01T10:00:00'
    },
    {
      id: 2,
      content: 'Learning something new everyday 💡 #coding',
      likes: 28,
      comments: 5,
      timestamp: '2024-02-28T15:30:00'
    }
  ]);
  const [newPost, setNewPost] = useState('');

  const handleFollow = () => {
    // TODO: Implement follow logic
  };

  const handleCreatePost = (e: React.FormEvent) => {
    e.preventDefault();
    if (newPost.trim()) {
      setPosts([{
        id: posts.length + 1,
        content: newPost,
        likes: 0,
        comments: 0,
        timestamp: new Date().toISOString()
      }, ...posts]);
      setNewPost('');
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6 bg-gray-900 min-h-screen text-gray-100">
      {/* Profile Header */}
      <div className="bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
        <div className="flex items-start gap-6">
          <img 
            src={dummyUser.avatar_url || "https://ui-avatars.com/api/?name=John+Doe&background=3b82f6&color=fff&bold=true"} 
            alt="Avatar" 
            className="w-32 h-32 rounded-full border-4 border-purple-100"
          />
          <div className="flex-1">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h1 className="text-3xl font-bold text-gray-100">
                  {dummyUser.first_name} {dummyUser.last_name}
                </h1>
                <p className="text-gray-400">@{dummyUser.username}</p>
              </div>
              <div className="flex items-center gap-4">
                <button 
                  onClick={handleFollow}
                  className="flex items-center gap-2 bg-purple-700 text-gray-100 px-6 py-2 rounded-full hover:bg-purple-600 transition-colors"
                >
                  <FiUsers className="text-lg" />
                  Follow
                </button>
                <button className="text-purple-400 hover:text-purple-300">
                  <FiEdit className="text-2xl" />
                </button>
              </div>
            </div>
            
            <p className="text-gray-300 mb-4">{dummyUser.about_me}</p>
            
            <div className="flex items-center gap-6 text-gray-400">
              <div className="flex items-center gap-2">
                <FiUsers />
                <span>1.2k followers</span>
              </div>
              <div className="flex items-center gap-2">
                <FiUsers />
                <span>856 following</span>
              </div>
              <button 
                onClick={() => setIsPublic(!isPublic)}
                className="flex items-center gap-2 ml-auto text-sm px-4 py-2 rounded-full bg-gray-700 hover:bg-gray-600 text-gray-200"
              >
                {isPublic ? <FiUnlock /> : <FiLock />}
                {isPublic ? 'Public Profile' : 'Private Profile'}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-gray-700 mb-6">
        <button
          className={`px-6 py-3 flex items-center gap-2 ${
            activeTab === 'posts' 
              ? 'border-b-2 border-purple-500 text-purple-400' 
              : 'text-gray-400 hover:text-purple-400'
          }`}
          onClick={() => setActiveTab('posts')}
        >
          <FiMessageSquare />
          Posts
        </button>
        <button
          className={`px-6 py-3 flex items-center gap-2 ${
            activeTab === 'followers' 
              ? 'border-b-2 border-purple-500 text-purple-400' 
              : 'text-gray-400 hover:text-purple-400'
          }`}
          onClick={() => setActiveTab('followers')}
        >
          <FiUsers />
          Followers
        </button>
      </div>

      {/* Content Area */}
      {activeTab === 'posts' ? (
        <div>
          {/* Create Post */}
          <form onSubmit={handleCreatePost} className="mb-6 bg-gray-800 rounded-lg shadow p-4">
            <textarea
              value={newPost}
              onChange={(e) => setNewPost(e.target.value)}
              placeholder="What's on your mind?"
              className="w-full p-4 border border-gray-700 bg-gray-900 text-gray-100 rounded-lg mb-4 resize-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              rows={3}
            />
            <div className="flex justify-between items-center">
              <button 
                type="button"
                className="text-purple-400 hover:text-purple-300 flex items-center gap-2"
              >
                <FiPlus />
                Add Image
              </button>
              <button 
                type="submit"
                className="bg-purple-700 text-gray-100 px-6 py-2 rounded-full hover:bg-purple-600 transition-colors"
              >
                Post
              </button>
            </div>
          </form>

          {/* Posts List */}
          {posts.map(post => (
            <div key={post.id} className="bg-gray-800 rounded-lg shadow p-6 mb-4 hover:bg-gray-750 transition-colors">
              <div className="flex items-center gap-4 mb-4">
                <img 
                  src={dummyUser.avatar_url || "https://ui-avatars.com/api/?name=John+Doe&background=3b82f6&color=fff&bold=true"} 
                  alt="Avatar" 
                  className="w-12 h-12 rounded-full border-2 border-gray-700"
                />
                <div>
                  <h3 className="font-semibold text-gray-100">{dummyUser.first_name} {dummyUser.last_name}</h3>
                  <p className="text-sm text-gray-400">
                    {new Date(post.timestamp).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <p className="text-gray-200 mb-4">{post.content}</p>
              <div className="flex items-center gap-6 text-gray-400">
                <button className="flex items-center gap-2 hover:text-purple-400">
                  <FiHeart /> {post.likes}
                </button>
                <button className="flex items-center gap-2 hover:text-purple-400">
                  <FiMessageSquare /> {post.comments}
                </button>
                <button className="flex items-center gap-2 hover:text-purple-400">
                  <FiShare />
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="bg-gray-800 rounded-lg shadow p-6">
          <div className="grid grid-cols-2 gap-4">
            {/* Followers List */}
            {[1, 2, 3, 4].map((follower) => (
              <div key={follower} className="flex items-center gap-4 p-4 hover:bg-gray-750 rounded-lg transition-colors">
                <img 
                  src="https://ui-avatars.com/api/?name=f+l&background=3b82f6&color=fff&bold=true" 
                  alt="Follower" 
                  className="w-12 h-12 rounded-full"
                />
                <div>
                  <h3 className="font-semibold text-gray-100">Follower {follower}</h3>
                  <p className="text-gray-400">@follower{follower}</p>
                </div>
                <button className="ml-auto text-sm bg-purple-700 text-gray-100 px-4 py-2 rounded-full hover:bg-purple-600">
                  Following
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
