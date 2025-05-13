import Link from 'next/link';
import { FiShare } from 'react-icons/fi';
import { Post } from '@/types/Post';
import { formatDistanceToNow } from 'date-fns';
import { Avatar } from '@/components/ui/avatar';
import Image from 'next/image'; // Add this import

interface PostCardProps {
  post: Post;
}

export default function PostCard({ post }: PostCardProps) {
  // const initials are not used with Avatar src logic change, can be removed if not needed elsewhere
  // const initials = post.user_first_name && post.user_last_name
  //   ? `${post.user_first_name[0]}${post.user_last_name[0]}`
  //   : undefined;

  return (
    <div className="bg-gray-800 rounded-lg shadow p-6 mb-4 hover:bg-gray-750 transition-colors">
      <div className="flex items-center gap-4 mb-4">
        <Link href={`/profile/${post.user_id}`} className="flex items-center gap-4">
          <Avatar
            src={post.user_avatar_url || `https://ui-avatars.com/api/?name=${post.user_first_name}+${post.user_last_name}&background=3b82f6&color=fff&bold=true`}
            alt={`${post.user_first_name || 'Anonymous'} ${post.user_last_name || 'User'}`}
            className="w-12 h-12 border-2 border-gray-700"
          />
          <div>
            <h3 className="font-semibold text-gray-100">
              {post.user_first_name || 'Anonymous'} {post.user_last_name || 'User'}
            </h3>
            <p className="text-sm text-gray-400">
              {formatDistanceToNow(new Date(post.created_at), { addSuffix: true })}
            </p>
          </div>
        </Link>
      </div>
      <Link href={`/posts/${post.id}`} className="block space-y-4">
        <h2 className="text-xl font-semibold text-gray-100">{post.title}</h2>
        <p className="text-gray-200">{post.content}</p>
        {post.image_url && (
          <div className="my-3 relative" style={{ paddingTop: '56.25%' /* 16:9 Aspect Ratio for responsive height */ }}>
            <Image
              src={post.image_url}
              alt={post.title || "Post image"}
              layout="fill"
              objectFit="cover"
              className="rounded-lg"
            />
          </div>
        )}
      </Link>
      <div className="flex items-center gap-6 mt-4 text-gray-400">
        <div className="flex items-center gap-2">
          <button className="text-sm text-gray-500">
            {post.privacy === 'public' ? 'Public' : post.privacy === 'semi_private' ? 'Followers Only' : 'Private'}
          </button>
        </div>
        <button className="flex items-center gap-2 hover:text-purple-400">
          <FiShare />
        </button>
      </div>
    </div>
  );
}