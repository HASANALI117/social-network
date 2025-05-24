"use client"

import { useEffect } from 'react'
import { Avatar } from "@/components/ui/avatar"
import {
  Sidebar,
  SidebarBody,
  SidebarFooter,
  SidebarHeader,
  SidebarHeading,
  SidebarItem,
  SidebarLabel,
  SidebarSection,
} from "@/components/ui/sidebar"
import {
  Dropdown,
  DropdownButton,
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '@/components/ui/dropdown'
import {
  BellIcon,
  ChatBubbleLeftIcon,
  ChevronUpIcon,
  PlusIcon,
  UsersIcon,
} from '@heroicons/react/16/solid'
import {
  HomeIcon,
  PhotoIcon,
  UserCircleIcon,
  UserGroupIcon,
  UsersIcon as UsersIconOutline,
} from '@heroicons/react/24/solid'
import { usePathname } from 'next/navigation'
import { AccountDropdownMenu } from './account-dropdown-menu'
import { NotificationsDropdown } from './notifications-dropdown'
import { Link } from "../ui/link"
import { ChevronDownIcon, Cog8ToothIcon } from "@heroicons/react/20/solid"
import { useUserStore } from '@/store/useUserStore'
import { useNotifications } from '@/hooks/useNotifications'
import { Notification } from '@/types/Notification'

export function AppSidebar() {
  const pathname = usePathname()
  const { user, isAuthenticated } = useUserStore()
  const {
    notifications,
    unreadCount,
    fetchNotifications,
    refreshNotifications,
    markAsRead,
    markAllAsRead
  } = useNotifications()

  useEffect(() => {
    useUserStore.persist.rehydrate()
  }, [])

  const handleNotificationClick = async (notification: Notification) => {
    if (!notification.is_read) {
      await markAsRead(notification.id)
      // Refresh notifications to ensure UI is updated
      refreshNotifications();
    }
    // Navigation is handled within NotificationsDropdown
  }

  const handleMarkAllNotificationsAsRead = async () => {
    await markAllAsRead()
    // Refresh notifications to ensure UI is updated
    refreshNotifications();
  }

  if (!isAuthenticated) {
    return (
      <Sidebar>
        <SidebarHeader>
          <SidebarItem href="/">
            <Avatar src="https://ui-avatars.com/api/?name=Social+Network&background=6366f1&color=fff&bold=true" className="bg-indigo-500" />
            <SidebarLabel>Social Network</SidebarLabel>
          </SidebarItem>
        </SidebarHeader>
        <SidebarBody>
          <SidebarSection>
            <SidebarItem href="/login">
              <UserCircleIcon />
              <SidebarLabel>Sign In</SidebarLabel>
            </SidebarItem>
            <SidebarItem href="/register">
              <UsersIconOutline />
              <SidebarLabel>Register</SidebarLabel>
            </SidebarItem>
          </SidebarSection>
        </SidebarBody>
      </Sidebar>
    )
  }

  return (
    <Sidebar>
      <SidebarHeader>
        <SidebarItem href="/">
          <Avatar src="https://ui-avatars.com/api/?name=Social+Network&background=6366f1&color=fff&bold=true" className="bg-indigo-500" />
          <SidebarLabel>Social Network</SidebarLabel>
        </SidebarItem>
      </SidebarHeader>

      <SidebarBody>
        <SidebarSection>
          <SidebarItem href="/feed" current={pathname === '/'}>
            <HomeIcon />
            <SidebarLabel>Feed</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/profile" current={pathname.startsWith('/profile')}>
            <UserCircleIcon />
            <SidebarLabel>Profile</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/groups" current={pathname.startsWith('/groups')}>
            <UserGroupIcon />
            <SidebarLabel>Groups</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/friends" current={pathname.startsWith('/friends')}>
            <UsersIconOutline />
            <SidebarLabel>Friends</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/posts" current={pathname.startsWith('/posts')}>
            <PhotoIcon />
            <SidebarLabel>Posts</SidebarLabel>
          </SidebarItem>
        </SidebarSection>

        <SidebarSection>
          <SidebarHeading>Communication</SidebarHeading>
          <Dropdown>
            <DropdownButton as={SidebarItem} className="relative">
              <BellIcon />
              <SidebarLabel>Notifications</SidebarLabel>
              {unreadCount > 0 && (
                <span className="absolute right-2 top-1/2 -translate-y-1/2 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-xs text-white">
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
          <SidebarItem href="/messages" className="relative">
            <ChatBubbleLeftIcon />
            <SidebarLabel>Messages</SidebarLabel>
            <span className="absolute right-2 top-1/2 -translate-y-1/2 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-xs text-white">
              2
            </span>
          </SidebarItem>
        </SidebarSection>

        <SidebarSection className="max-lg:hidden">
          <SidebarHeading>Groups</SidebarHeading>
          <SidebarItem href="/groups/create">
            <PlusIcon />
            <SidebarLabel>Create New Group</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/groups/my-groups">
            <UsersIcon />
            <SidebarLabel>My Groups</SidebarLabel>
          </SidebarItem>
        </SidebarSection>
      </SidebarBody>

      <SidebarFooter className="max-lg:hidden">
        {user && (
          <Dropdown>
            <DropdownButton as={SidebarItem}>
              <span className="flex min-w-0 items-center gap-3">
                <Avatar 
                  src={user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff&bold=true`} 
                  className="size-10" 
                  square 
                  alt={`${user.first_name} ${user.last_name}`} 
                />
                <span className="min-w-0">
                  <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">
                    {user.first_name}
                  </span>
                  <span className="block truncate text-xs/5 font-normal text-zinc-500 dark:text-zinc-400">
                    {user.email}
                  </span>
                </span>
              </span>
              <ChevronUpIcon />
            </DropdownButton>
            <AccountDropdownMenu anchor="top start" />
          </Dropdown>
        )}
      </SidebarFooter>
    </Sidebar>
  )
}
