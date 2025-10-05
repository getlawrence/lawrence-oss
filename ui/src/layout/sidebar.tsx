import {
  Settings,
  Server,
  Users,
  FileText,
  BarChart3,
  GitBranch,
  Edit3,
  Sparkle,
  ChevronDown,
} from 'lucide-react';
import * as React from 'react';
import { Link, useLocation } from 'react-router-dom';

import { ModeToggle } from './mode-toggle';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarFooter,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
  useSidebar,
} from '@/components/ui/sidebar';

interface MenuItem {
  key: string;
  title: string;
  url: string;
  icon: React.ComponentType<{ className?: string }>;
}

export function AppSidebar() {
  const location = useLocation();
  const { state } = useSidebar();

  const mainItems: MenuItem[] = [
    {
      key: 'inventory',
      title: 'Inventory',
      url: '/inventory',
      icon: Server,
    },
    {
      key: 'topology',
      title: 'Topology',
      url: '/topology',
      icon: GitBranch,
    },
    {
      key: 'groups',
      title: 'Groups',
      url: '/groups',
      icon: Users,
    },
    {
      key: 'configs',
      title: 'Configs',
      url: '/configs',
      icon: FileText,
    },
    {
      key: 'config-editor',
      title: 'Config Editor',
      url: '/config-editor',
      icon: Edit3,
    },
    {
      key: 'telemetry',
      title: 'Telemetry',
      url: '/telemetry',
      icon: BarChart3,
    },
  ];

  return (
    <Sidebar collapsible="icon" className="border-r border-border">
      <SidebarHeader className="border-b border-border h-16 flex items-center justify-center relative">
        <SidebarMenu>
          <SidebarMenuItem>
            {state === 'collapsed' ? (
              <div className="relative group">
                <div className="flex items-center justify-center h-8 w-8 rounded-md transition-colors group-hover:opacity-0">
                  <Sparkle className="h-4 w-4 text-primary" />
                </div>
                <SidebarTrigger className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity" />
              </div>
            ) : (
              <SidebarMenuButton asChild>
                <a href="/" className="flex items-center space-x-2">
                  <Sparkle className="h-4 w-4 text-primary" />
                  <span>Lawrence OSS</span>
                </a>
              </SidebarMenuButton>
            )}
          </SidebarMenuItem>
        </SidebarMenu>
        {state === 'expanded' && (
          <SidebarTrigger className="absolute right-2 top-1/2 -translate-y-1/2 h-6 w-6" />
        )}
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarMenu>
            {mainItems.map(item => {
              const isActive = location.pathname === item.url;
              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild isActive={isActive} tooltip={item.title}>
                    <Link to={item.url} className="relative">
                      <item.icon />
                      {state === 'expanded' && <span>{item.title}</span>}
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              );
            })}
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t border-border">
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton>
                  {state === 'expanded' ? (
                    <>
                      <div className="flex items-center gap-2 flex-1 min-w-0">
                        <div className="w-5 h-5 bg-muted rounded-full flex items-center justify-center flex-shrink-0">
                          <span className="text-xs font-medium">L</span>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="truncate text-sm font-medium">Lawrence OSS</div>
                          <div className="truncate text-xs text-muted-foreground">
                            Version 1.0.0
                          </div>
                        </div>
                      </div>
                      <ChevronDown className="ml-auto" />
                    </>
                  ) : (
                    <>
                      <div className="w-6 h-6 bg-primary/10 rounded-full flex items-center justify-center">
                        <span className="text-xs font-medium text-primary">L</span>
                      </div>
                      <ChevronDown className="ml-auto" />
                    </>
                  )}
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-[calc(var(--radix-popper-anchor-width)-1rem)]">
                <DropdownMenuItem asChild>
                  <ModeToggle />
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}
