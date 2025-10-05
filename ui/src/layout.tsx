import { Outlet } from 'react-router-dom';

import { AppSidebar } from './layout/sidebar';

import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { getCookie } from '@/lib/utils';

export default function Layout() {
  // Read the sidebar state from cookie, default to true (expanded) if not found
  const sidebarCookie = getCookie('sidebar_state');
  const defaultOpen = sidebarCookie === null ? true : sidebarCookie === 'true';

  return (
    <SidebarProvider defaultOpen={defaultOpen} className="flex h-screen w-screen overflow-hidden">
      <AppSidebar />
      <SidebarInset className="flex-1">
        <div className="flex flex-1 flex-col gap-4 p-4 w-full h-screen min-h-0 overflow-auto">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
