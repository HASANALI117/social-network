import Link from "next/link";
import Image from "next/image";
import { UserCircleIcon } from "@heroicons/react/24/solid";
import { Heading } from "@/components/ui/heading";
import { Text } from "@/components/ui/text";
import { Group } from "@/types/Group";

interface GroupCardProps {
  group: Group;
}

// A simple date formatter, replace with a more robust one if available
const formatDate = (dateString: string) => {
  if (!dateString) return "N/A";
  try {
    return new Date(dateString).toLocaleDateString();
  } catch (error) {
    console.error("Error formatting date:", error);
    return dateString; // return original if formatting fails
  }
};

const GroupCard: React.FC<GroupCardProps> = ({ group }) => {
  return (
    <Link key={group.id} href={`/groups/${group.id}`} passHref>
      <div className="bg-gray-800 p-4 sm:p-6 rounded-lg shadow-lg hover:bg-gray-700 transition-colors cursor-pointer flex flex-col sm:flex-row gap-4 items-start mb-4">
        {group.avatar_url ? (
          <img
            src={group.avatar_url}
            alt={`${group.name} avatar`}
            className="rounded-md object-cover w-20 h-20 flex-shrink-0"
          />
        ) : (
          <div className="w-20 h-20 bg-gray-700 rounded-md flex items-center justify-center flex-shrink-0">
            <UserCircleIcon className="h-12 w-12 text-gray-500" />
          </div>
        )}
        <div className="flex-grow">
          <Heading level={3} className="mb-1 truncate">
            {group.name}
          </Heading>
          <Text className="text-gray-400 text-sm line-clamp-3 mb-2">
            {group.description || "No description provided."}
          </Text>
          <div className="text-xs text-gray-500 space-y-1">
            {/* Creator info can be added back if needed, requires group.creator_info */}
            {/* <Text>Created by: {group.creator_info ? `${group.creator_info.first_name} ${group.creator_info.last_name}` : 'N/A'}</Text> */}
            <Text>
              Members: {group.members_count ?? 0} | Posts:{" "}
              {group.posts_count ?? 0} | Events: {group.events_count ?? 0}
            </Text>
            <Text>Created: {formatDate(group.created_at)}</Text>
          </div>
        </div>
      </div>
    </Link>
  );
};

export default GroupCard;
