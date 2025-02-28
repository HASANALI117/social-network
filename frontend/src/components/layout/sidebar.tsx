import { Avatar } from "@/components/ui/avatar"
import {
  Sidebar,
  SidebarBody,
  SidebarHeader,
  SidebarSection,
  SidebarItem,
  SidebarLabel,
} from "@/components/ui/sidebar"
import {
  HomeIcon,
  UserIcon,
  UserGroupIcon,
  ChatBubbleLeftIcon,
  BellIcon,
  UsersIcon,
  Cog6ToothIcon,
} from "@heroicons/react/24/outline"

export function AppSidebar() {
  return (
    <Sidebar>
      <SidebarHeader>
        <SidebarItem>
          <Avatar src="/vercel.svg" />
          <SidebarLabel>Social Network</SidebarLabel>
        </SidebarItem>
      </SidebarHeader>
      <SidebarBody>
        <SidebarSection>
          <SidebarItem href="/">
            <HomeIcon className="h-6 w-6" />
            <SidebarLabel>Home</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/profile">
            <UserIcon className="h-6 w-6" />
            <SidebarLabel>Profile</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/groups">
            <UserGroupIcon className="h-6 w-6" />
            <SidebarLabel>Groups</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/chat">
            <ChatBubbleLeftIcon className="h-6 w-6" />
            <SidebarLabel>Chat</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/notifications">
            <BellIcon className="h-6 w-6" />
            <SidebarLabel>Notifications</SidebarLabel>
          </SidebarItem>
          <SidebarItem href="/followers">
            <UsersIcon className="h-6 w-6" />
            <SidebarLabel>Followers</SidebarLabel>
          </SidebarItem>
        </SidebarSection>
        <SidebarSection>
          <SidebarItem href="/settings">
            <Cog6ToothIcon className="h-6 w-6" />
            <SidebarLabel>Settings</SidebarLabel>
          </SidebarItem>
        </SidebarSection>
      </SidebarBody>
    </Sidebar>
  )
}
