import { FiMessageSquare, FiUsers } from 'react-icons/fi';

interface TabSwitcherProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

export default function TabSwitcher({ activeTab, onTabChange }: TabSwitcherProps) {
  return (
    <div className="flex border-b border-gray-700 mb-6">
      <button
        className={`px-6 py-3 flex items-center gap-2 ${
          activeTab === 'posts' 
            ? 'border-b-2 border-purple-500 text-purple-400' 
            : 'text-gray-400 hover:text-purple-400'
        }`}
        onClick={() => onTabChange('posts')}
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
        onClick={() => onTabChange('followers')}
      >
        <FiUsers />
        Followers
      </button>
    </div>
  );
}
