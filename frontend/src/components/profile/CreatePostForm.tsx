import { FiPlus } from 'react-icons/fi';
import { useState, FormEvent } from 'react';

interface CreatePostFormProps {
  onSubmit: (content: string) => void;
}

export default function CreatePostForm({ onSubmit }: CreatePostFormProps) {
  const [newPost, setNewPost] = useState('');

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (newPost.trim()) {
      onSubmit(newPost);
      setNewPost('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="mb-6 bg-gray-800 rounded-lg shadow p-4">
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
  );
}
