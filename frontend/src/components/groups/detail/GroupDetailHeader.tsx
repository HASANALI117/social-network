import React from 'react';
import Link from 'next/link';
import { Avatar } from '../../ui/avatar';
import { Heading } from '../../ui/heading';
import { Text } from '../../ui/text';
import { UserBasicInfo } from '../../../types/User';

interface GroupDetailHeaderProps {
  name: string;
  description: string;
  avatar_url: string | null | undefined;
  creator_info: UserBasicInfo;
  created_at: string | Date | undefined;
  groupId: string;
  isMember: boolean;
}

export default function GroupDetailHeader({
  name,
  description,
  avatar_url,
  creator_info,
  created_at,
  groupId,
  isMember,
}: GroupDetailHeaderProps) {
  const creatorFullName = `${creator_info.first_name} ${creator_info.last_name}`;
  const creatorUsername = creator_info.username;
  const creatorAvatarUrl = creator_info.avatar_url;

  const formatDate = (dateInput: string | Date | undefined): string => {
    if (!dateInput) return 'Date not available';
    try {
      const dateObj = new Date(dateInput);
      if (!isNaN(dateObj.getTime())) {
        return dateObj.toLocaleDateString();
      } else {
        console.error("Invalid date string for group created_at:", dateInput);
        return 'Invalid date';
      }
    } catch (e) {
      console.error("Error parsing date string for group created_at:", dateInput, e);
      return 'Error parsing date';
    }
  };

  return (
    <header className="mb-8 p-6 bg-gray-800 rounded-lg shadow-xl">
      <div className="flex flex-col sm:flex-row items-center">
        <Avatar
          src={avatar_url || null}
          initials={!avatar_url && name ? name.substring(0, 1).toUpperCase() : undefined}
          alt={`${name} avatar`}
          className="h-24 w-24 md:h-32 md:w-32 rounded-full mr-0 sm:mr-6 mb-4 sm:mb-0 border-2 border-purple-500"
        />
        <div className="text-center sm:text-left flex-grow">
          <Heading level={1} className="text-3xl md:text-4xl font-bold text-purple-400 mb-2">
            {name}
          </Heading>
          <Text className="text-gray-300 mb-3 text-lg">{description}</Text>
          <div className="flex items-center justify-center sm:justify-start text-sm text-gray-400">
            <Avatar
              src={creatorAvatarUrl || null}
              initials={!creatorAvatarUrl && creatorFullName ? creatorFullName.substring(0, 1).toUpperCase() : undefined}
              alt={creatorFullName}
              className="h-8 w-8 mr-2 rounded-full border border-gray-600"
            />
            <Text>
              Created by:{' '}
              <Link href={`/profile/${creator_info.user_id}`} className="text-purple-300 hover:underline">
                {creatorFullName} ({creatorUsername})
              </Link>
            </Text>
            <span className="mx-2 text-gray-500">|</span>
            <Text>Created on: {formatDate(created_at)}</Text>
          </div>
        </div>
        {isMember && (
          <div className="mt-4 sm:mt-0 sm:ml-4">
            <Link
              href={`/groups/${groupId}/chat`}
              className="inline-flex items-center px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors"
            >
              Chat
            </Link>
          </div>
        )}
      </div>
    </header>
  );
}