"use client"

import { Avatar } from "@/components/ui/avatar"
import {
  Dropdown,
  DropdownButton,
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from "@/components/ui/dropdown"
import { Navbar, NavbarItem, NavbarSection, NavbarSpacer } from "@/components/ui/navbar"
import { BellIcon, ChatBubbleLeftIcon, UserGroupIcon } from '@heroicons/react/16/solid'
import { AccountDropdownMenu } from './account-dropdown-menu'
import { NotificationsDropdown } from './notifications-dropdown'

export function AppNavbar() {
  return (
    <Navbar>
      <NavbarSection className="lg:hidden">
        <Dropdown>
          <DropdownButton as={NavbarItem} className="relative">
            <BellIcon className="h-6 w-6" />
            <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-xs text-white">
              3
            </span>
          </DropdownButton>
          <NotificationsDropdown anchor="bottom start" />
        </Dropdown>
        <NavbarItem href="/messages" className="relative">
          <ChatBubbleLeftIcon className="h-6 w-6" />
          <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-xs text-white">
            2
          </span>
        </NavbarItem>
        <NavbarItem href="/groups">
          <UserGroupIcon className="h-6 w-6" />
        </NavbarItem>
      </NavbarSection>
      <NavbarSpacer />
      <NavbarSection>
        <Dropdown>
          <DropdownButton as={NavbarItem}>
            <Avatar src="https://ui-avatars.com/api/?name=Erica+Jones&background=3b82f6&color=fff&bold=true" square alt="Erica Jones" />
          </DropdownButton>
          <AccountDropdownMenu anchor="bottom end" />
        </Dropdown>
      </NavbarSection>
    </Navbar>
  )
}
