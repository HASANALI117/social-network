"use client"

import {
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '@/components/ui/dropdown'
import {
  ArrowRightStartOnRectangleIcon,
  LightBulbIcon,
  ShieldCheckIcon,
  UserCircleIcon,
} from '@heroicons/react/16/solid'
import { useUserStore } from '@/store/useUserStore'
import { useRouter } from 'next/navigation'
import { Avatar } from '../ui/avatar'

export function AccountDropdownMenu({ anchor }: { anchor: 'top start' | 'bottom end' }) {
  const { user, logout } = useUserStore()
  const router = useRouter()

  const handleLogout = () => {
    logout()
    router.push('/login')
  }

  if (!user) return null

  return (
    <DropdownMenu className="min-w-64" anchor={anchor}>
      <div className="px-4 py-3">
        <div className="flex items-center space-x-3">
          <Avatar 
            src={user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff&bold=true`}
            square 
            alt={`${user.first_name} ${user.last_name}`} 
          />
          <div>
            <div className="font-medium">{`${user.first_name} ${user.last_name}`}</div>
            <div className="text-sm text-gray-500">{user.email}</div>
          </div>
        </div>
      </div>
      <DropdownDivider />
      <DropdownItem href="/profile">
        <UserCircleIcon />
        <DropdownLabel>My profile</DropdownLabel>
      </DropdownItem>
      <DropdownDivider />
      <DropdownItem href="/privacy">
        <ShieldCheckIcon />
        <DropdownLabel>Privacy settings</DropdownLabel>
      </DropdownItem>
      <DropdownItem href="/feedback">
        <LightBulbIcon />
        <DropdownLabel>Share feedback</DropdownLabel>
      </DropdownItem>
      <DropdownDivider />
      <DropdownItem onClick={handleLogout}>
        <ArrowRightStartOnRectangleIcon />
        <DropdownLabel>Sign out</DropdownLabel>
      </DropdownItem>
    </DropdownMenu>
  )
}
