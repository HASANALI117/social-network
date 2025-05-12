import { User } from '@/types/User';
import UserCard from './UserCard';

interface UserListProps {
  users: User[];
  emptyMessage?: string;
}

export default function UserList({ users, emptyMessage = 'No users to display' }: UserListProps) {
  if (users.length === 0) {
    return (
      <div className="text-center text-gray-400 py-8">
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {users.map((user) => (
        <UserCard key={user.id} user={user} />
      ))}
    </div>
  );
}