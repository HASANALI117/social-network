import { FiEdit, FiUsers, FiLock, FiUnlock } from 'react-icons/fi';
import { User } from '@/types/User';

interface ProfileHeaderProps {
  user: User;
  isPublic: boolean;
  onTogglePublic: () => void;
  onFollow: () => void;
}

export default function ProfileHeader({ user, isPublic, onTogglePublic, onFollow }: ProfileHeaderProps) {
  return (
    <div className="bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
      <div className="flex items-start gap-6">
        <img 
          src={user.avatar_url || "https://ui-avatars.com/api/?name=John+Doe&background=3b82f6&color=fff&bold=true"} 
          alt="Avatar" 
          className="w-32 h-32 rounded-full border-4 border-purple-100"
        />
        <div className="flex-1">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-3xl font-bold text-gray-100">
                {user.first_name} {user.last_name}
              </h1>
              <p className="text-gray-400">@{user.username}</p>
            </div>
            <div className="flex items-center gap-4">
              <button 
                onClick={onFollow}
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
          
          <p className="text-gray-300 mb-4">{user.about_me}</p>
          
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
              onClick={onTogglePublic}
              className="flex items-center gap-2 ml-auto text-sm px-4 py-2 rounded-full bg-gray-700 hover:bg-gray-600 text-gray-200"
            >
              {isPublic ? <FiUnlock /> : <FiLock />}
              {isPublic ? 'Public Profile' : 'Private Profile'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
