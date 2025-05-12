import Link from 'next/link';
import { User } from '@/types/User';

interface UserCardProps {
  user: User;
}

export default function UserCard({ user }: UserCardProps) {
  return (
    <Link 
      href={`/profile/${user.id}`}
      className="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-800 transition-colors"
    >
      <img
        src={user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff&bold=true`}
        alt={`${user.first_name} ${user.last_name}`}
        className="w-10 h-10 rounded-full"
      />
      <div>
        <div className="font-medium text-gray-100">
          {user.first_name} {user.last_name}
        </div>
        <div className="text-sm text-gray-400">@{user.username}</div>
      </div>
    </Link>
  );
}