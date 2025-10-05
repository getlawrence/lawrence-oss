import { Outlet } from 'react-router-dom';
import { Sidebar } from './layout/sidebar';

export default function Layout() {
    return (
        <div className="flex h-screen bg-gray-50">
            <Sidebar />
            <main className="flex-1 overflow-y-auto">
                <Outlet />
            </main>
        </div>
    );
}
