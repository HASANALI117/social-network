import { FiShare } from 'react-icons/fi';
import { Post } from '@/types/Post';
import { formatDistanceToNow } from 'date-fns';

interface DummyUser {
  avatar_url: string;
  first_name: string;
  last_name: string;
}

interface PostCardProps {
  post: Post;
}

export default function PostCard({ post }: PostCardProps) {
  // Dummy user info until we have real user data
  const dummyUser: DummyUser = {
    avatar_url: 'https://ui-avatars.com/api/?name=John+Doe&background=3b82f6&color=fff&bold=true',
    first_name: 'John',
    last_name: 'Doe'
  };

  return (
    <div className="bg-gray-800 rounded-lg shadow p-6 mb-4 hover:bg-gray-750 transition-colors">
      <div className="flex items-center gap-4 mb-4">
        <img
          src={dummyUser.avatar_url}
          alt="Avatar"
          className="w-12 h-12 rounded-full border-2 border-gray-700"
        />
        <div>
          <h3 className="font-semibold text-gray-100">
            {dummyUser.first_name} {dummyUser.last_name}
          </h3>
          <p className="text-sm text-gray-400">
            {formatDistanceToNow(post.createdAt, { addSuffix: true })}
          </p>
        </div>
      </div>
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-gray-100">{post.title}</h2>
        <p className="text-gray-200">{post.content}</p>
        {post.image_url && (
          <img
            src={post.image_url}
            alt={post.title}
            className="w-full rounded-lg"
          />
        )}
      </div>
      <div className="flex items-center gap-6 mt-4 text-gray-400">
        <div className="flex items-center gap-2">
          <button className="text-sm text-gray-500">
            {post.privacy === 'public' ? 'Public' : post.privacy === 'friends' ? 'Friends Only' : 'Private'}
          </button>
        </div>
        <button className="flex items-center gap-2 hover:text-purple-400">
          <FiShare />
        </button>
      </div>
    </div>
  );
}
