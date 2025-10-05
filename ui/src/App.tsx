import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

import Layout from './layout';
import InventoryPage from './pages/Inventory';
import GroupsPage from './pages/Groups';
import ConfigsPage from './pages/Configs';
import TelemetryPage from './pages/Telemetry';

import './App.css';
import { ThemeProvider } from '@/components/ThemeProvider';
import { SWRProvider } from '@/lib/swr-provider';
import { ApiProvider } from '@/providers/ApiProvider';

function App() {
  return (
    <ThemeProvider defaultTheme="system">
      <SWRProvider>
        <ApiProvider>
          <Router>
            <Routes>
              {/* Main application routes */}
              <Route
                element={<Layout />}
              >
                <Route path="/" element={<InventoryPage />} />
                <Route path="/inventory" element={<InventoryPage />} />
                <Route path="/groups" element={<GroupsPage />} />
                <Route path="/configs" element={<ConfigsPage />} />
                <Route path="/telemetry" element={<TelemetryPage />} />
              </Route>
            </Routes>
          </Router>
        </ApiProvider>
      </SWRProvider>
    </ThemeProvider>
  );
}

export default App;