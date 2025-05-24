'use client';

import { useEffect } from 'react';
import { Avatar } from '@/components/ui/avatar';
import {
  Dropdown,
  DropdownButton,
  // DropdownDivider, // Not directly used here
  // DropdownItem, // Not directly used here
  // DropdownLabel, // Not directly used here
  // DropdownMenu, // Not directly used here
} from '@/components/ui/dropdown';
import {
  Navbar,
  NavbarItem,
  NavbarSection,
  NavbarSpacer,
} from '@/components/ui/navbar';
import {
  BellIcon,
  ChatBubbleLeftIcon,
  UserGroupIcon,
} from '@heroicons/react/16/solid';
import { AccountDropdownMenu } from './account-dropdown-menu';
import { NotificationsDropdown } from './notifications-dropdown';
import { Notification } from '@/types/Notification'; // Import Notification type
import { useUserStore } from '@/store/useUserStore';
import { useNotifications } from '@/hooks/useNotifications'; // Import the new hook
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';
// import Link from 'next/link' // Not directly used here

export function AppNavbar() {
  const { user, isAuthenticated } = useUserStore();
  const {
    notifications,
    unreadCount,
    fetchNotifications,
    refreshNotifications,
    markAsRead,
    markAllAsRead,
  } = useNotifications();
  const { messageCount } = useGlobalWebSocket();

  useEffect(() => {
    // Zustand's persist middleware handles rehydration automatically if configured correctly.
    // Explicit rehydration call might not be needed here if skipHydration is false or onRehydrateStorage is used.
    // useUserStore.persist.rehydrate()
  }, []);

  const avatarUrl =
    user?.avatar_url ||
    `https://ui-avatars.com/api/?name=${user?.first_name}+${user?.last_name}&background=3b82f6&color=fff&bold=true`;

  const handleNotificationClick = async (notification: Notification) => {
    if (!notification.is_read) {
      await markAsRead(notification.id);
      // Refresh notifications to ensure UI is updated
      refreshNotifications();
    }
    // Navigation is handled within NotificationsDropdown for now
  };

  const handleMarkAllNotificationsAsRead = async () => {
    await markAllAsRead();
    // Refresh notifications to ensure UI is updated
    refreshNotifications();
  };

  if (!isAuthenticated) {
    return (
      <Navbar>
        <NavbarSpacer />
        <NavbarSection>
          <NavbarItem href="/login">Sign In</NavbarItem>
          <NavbarItem href="/register">Register</NavbarItem>
        </NavbarSection>
      </Navbar>
    );
  }

  return (
    <Navbar>
      <NavbarSection className="lg:hidden">
        {' '}
        {/* This section seems to be for mobile/smaller screens */}
        <Dropdown>
          <DropdownButton as={NavbarItem} className="relative">
            <BellIcon className="h-6 w-6" />
            {unreadCount > 0 && (
              <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-xs text-white">
                {unreadCount > 9 ? '9+' : unreadCount}
              </span>
            )}
          </DropdownButton>
          <NotificationsDropdown
            anchor="bottom start"
            notifications={notifications}
            unreadCount={unreadCount}
            onNotificationClick={handleNotificationClick}
            onMarkAllAsRead={handleMarkAllNotificationsAsRead}
            fetchNotifications={refreshNotifications}
          />
        </Dropdown>
        <NavbarItem href="/messages" className="relative">
          <ChatBubbleLeftIcon className="h-6 w-6" />
          {messageCount > 0 && (
            <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-xs text-white">
              {messageCount > 9 ? '9+' : messageCount}
            </span>
          )}
        </NavbarItem>
        <NavbarItem href="/groups">
          <UserGroupIcon className="h-6 w-6" />
        </NavbarItem>
      </NavbarSection>

      {/* Consider if notifications bell should also be visible on larger screens */}
      {/* Example: Adding it to the main NavbarSection for larger screens */}
      <NavbarSection className="hidden lg:flex items-center gap-x-3">
        <Dropdown>
          <DropdownButton as={NavbarItem} className="relative">
            <BellIcon className="h-6 w-6" />
            {unreadCount > 0 && (
              <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-xs text-white">
                {unreadCount > 9 ? '9+' : unreadCount}
              </span>
            )}
          </DropdownButton>
          <NotificationsDropdown
            anchor="bottom end" // Adjusted anchor for larger screens
            notifications={notifications}
            unreadCount={unreadCount}
            onNotificationClick={handleNotificationClick}
            onMarkAllAsRead={handleMarkAllNotificationsAsRead}
            fetchNotifications={refreshNotifications}
          />
        </Dropdown>
        <NavbarItem href="/messages" className="relative">
          <ChatBubbleLeftIcon className="h-6 w-6" />
          {messageCount > 0 && (
            <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-xs text-white">
              {messageCount > 9 ? '9+' : messageCount}
            </span>
          )}
        </NavbarItem>
        <NavbarItem href="/groups">
          <UserGroupIcon className="h-6 w-6" />
        </NavbarItem>
      </NavbarSection>

      <NavbarSpacer />
      <NavbarSection>
        <Dropdown>
          <DropdownButton as={NavbarItem}>
            <Avatar
              src={avatarUrl}
              square
              alt={`${user?.first_name} ${user?.last_name}`}
              className="size-8"
            />
          </DropdownButton>
          <AccountDropdownMenu anchor="bottom end" />
        </Dropdown>
      </NavbarSection>
    </Navbar>
  );
}
