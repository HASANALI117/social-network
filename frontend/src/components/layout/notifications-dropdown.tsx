"use client"

import {
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '@/components/ui/dropdown'
import { Avatar } from '@/components/ui/avatar'
import { UserPlusIcon, UserGroupIcon, CalendarDaysIcon } from '@heroicons/react/24/solid'

export function NotificationsDropdown({ anchor }: { anchor: 'top start' | 'bottom start' | 'bottom end' }) {
  return (
    <DropdownMenu className="min-w-80" anchor={anchor}>
      <DropdownItem href="#" className="flex items-start gap-3 py-2">
        <div className="flex-1 space-y-1">
          <DropdownLabel className="font-medium">Follow Request</DropdownLabel>
          <p className="text-sm text-zinc-600 dark:text-zinc-400">John Smith wants to follow you</p>
        </div>
        <UserPlusIcon className="h-5 w-5 text-blue-500" />
      </DropdownItem>
      
      <DropdownDivider />
      
      <DropdownItem href="#" className="flex items-start gap-3 py-2">
        <div className="flex-1 space-y-1">
          <DropdownLabel className="font-medium">Group Invitation</DropdownLabel>
          <p className="text-sm text-zinc-600 dark:text-zinc-400">Sarah invited you to join "Photography Club"</p>
        </div>
        <UserGroupIcon className="h-5 w-5 text-purple-500" />
      </DropdownItem>
      
      <DropdownDivider />
      
      <DropdownItem href="#" className="flex items-start gap-3 py-2">
        <div className="flex-1 space-y-1">
          <DropdownLabel className="font-medium">New Event</DropdownLabel>
          <p className="text-sm text-zinc-600 dark:text-zinc-400">"Summer Meetup" event was created in Catalyst group</p>
        </div>
        <CalendarDaysIcon className="h-5 w-5 text-green-500" />
      </DropdownItem>
    </DropdownMenu>
  )
}
