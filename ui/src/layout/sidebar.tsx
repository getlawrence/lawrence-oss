import { NavLink } from 'react-router-dom';
import { cn } from '@/lib/utils';
import {
  Server,
  Users,
  FileText,
  BarChart3,
  Settings
} from 'lucide-react';

const navigation = [
  { name: 'Inventory', href: '/inventory', icon: Server },
  { name: 'Groups', href: '/groups', icon: Users },
  { name: 'Configs', href: '/configs', icon: FileText },
  { name: 'Telemetry', href: '/telemetry', icon: BarChart3 },
];

export function Sidebar() {
  return (
    <div className="flex flex-col w-64 bg-white border-r border-gray-200">
      {/* Logo */}
      <div className="flex items-center px-6 py-4 border-b border-gray-200">
        <div className="flex items-center space-x-2">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
            <Server className="h-5 w-5 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-semibold text-gray-900">Lawrence</h1>
            <p className="text-xs text-gray-500">OSS Edition</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-4 space-y-1">
        {navigation.map((item) => (
          <NavLink
            key={item.name}
            to={item.href}
            className={({ isActive }) =>
              cn(
                'flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors',
                isActive
                  ? 'bg-blue-50 text-blue-700 border-r-2 border-blue-700'
                  : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
              )
            }
          >
            <item.icon className="mr-3 h-5 w-5" />
            {item.name}
          </NavLink>
        ))}
      </nav>

      {/* Footer */}
      <div className="px-4 py-4 border-t border-gray-200">
        <div className="flex items-center space-x-2 text-sm text-gray-500">
          <Settings className="h-4 w-4" />
          <span>Version 1.0.0</span>
        </div>
      </div>
    </div>
  );
}