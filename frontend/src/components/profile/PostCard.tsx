import { FiHeart, FiMessageSquare, FiShare } from 'react-icons/fi';
import { UserType } from '@/types/User';

interface Post {
  id: number;
  content: string;
  likes: number;
  comments: number;
  timestamp: string;
}

interface PostCardProps {
  post: Post;
  user: UserType;
}

export default function PostCard({ post, user }: PostCardProps) {
  return (
    <div className="bg-gray-800 rounded-lg shadow p-6 mb-4 hover:bg-gray-750 transition-colors">
      <div className="flex items-center gap-4 mb-4">
        <img
          src={
            user.avatar_url ||
            'https://ui-avatars.com/api/?name=John+Doe&background=3b82f6&color=fff&bold=true'
          }
          alt="Avatar"
          className="w-12 h-12 rounded-full border-2 border-gray-700"
        />
        <div>
          <h3 className="font-semibold text-gray-100">
            {user.first_name} {user.last_name}
          </h3>
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
  );
}
