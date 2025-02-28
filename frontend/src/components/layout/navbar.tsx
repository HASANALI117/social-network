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
import {
  ArrowRightStartOnRectangleIcon,
  Cog8ToothIcon,
  UserIcon,
  BellIcon,
  ChatBubbleLeftIcon,
  UserGroupIcon,
} from "@heroicons/react/24/outline"

export function AppNavbar() {
  return (
    <Navbar>
      <NavbarSpacer />
      <NavbarSection>
        <NavbarItem href="/chat" aria-label="Chat">
          <ChatBubbleLeftIcon className="h-6 w-6" />
        </NavbarItem>
        <NavbarItem href="/groups" aria-label="Groups">
          <UserGroupIcon className="h-6 w-6" />
        </NavbarItem>
        <Dropdown>
          <DropdownButton as={NavbarItem} aria-label="Notifications">
            <BellIcon className="h-6 w-6" />
          </DropdownButton>
          <DropdownMenu className="min-w-80" anchor="bottom end">
            <DropdownItem href="/notifications">
              <DropdownLabel>No new notifications</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
        <Dropdown>
          <DropdownButton as={NavbarItem}>
            <Avatar square initials="U" />
          </DropdownButton>
          <DropdownMenu className="min-w-64" anchor="bottom end">
            <DropdownItem href="/profile">
              <UserIcon className="h-5 w-5" />
              <DropdownLabel>Profile</DropdownLabel>
            </DropdownItem>
            <DropdownItem href="/settings">
              <Cog8ToothIcon className="h-5 w-5" />
              <DropdownLabel>Settings</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem href="/logout">
              <ArrowRightStartOnRectangleIcon className="h-5 w-5" />
              <DropdownLabel>Sign out</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </NavbarSection>
    </Navbar>
  )
}
