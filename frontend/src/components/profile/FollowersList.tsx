interface Follower {
  id: number;
  name: string;
  username: string;
  avatarUrl: string;
}

export default function FollowersList() {
  const dummyFollowers: Follower[] = [1, 2, 3, 4].map(id => ({
    id,
    name: `Follower ${id}`,
    username: `follower${id}`,
    avatarUrl: "https://ui-avatars.com/api/?name=f+l&background=3b82f6&color=fff&bold=true"
  }));

  return (
    <div className="bg-gray-800 rounded-lg shadow p-6">
      <div className="grid grid-cols-2 gap-4">
        {dummyFollowers.map((follower) => (
          <div key={follower.id} className="flex items-center gap-4 p-4 hover:bg-gray-750 rounded-lg transition-colors">
            <img 
              src={follower.avatarUrl}
              alt={follower.name} 
              className="w-12 h-12 rounded-full"
            />
            <div>
              <h3 className="font-semibold text-gray-100">{follower.name}</h3>
              <p className="text-gray-400">@{follower.username}</p>
            </div>
            <button className="ml-auto text-sm bg-purple-700 text-gray-100 px-4 py-2 rounded-full hover:bg-purple-600">
              Following
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
