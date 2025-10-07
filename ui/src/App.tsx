import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

import Layout from './layout';
import AgentsPage from './pages/Agents';
import GroupsPage from './pages/Groups';
import GroupDetailsPage from './pages/GroupDetails';
import ConfigsPage from './pages/Configs';
import ConfigEditorPage from './pages/ConfigEditor';
import TelemetryPage from './pages/Telemetry';
import TopologyPage from './pages/Topology';

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
                <Route path="/" element={<AgentsPage />} />
                <Route path="/agents" element={<AgentsPage />} />
                <Route path="/groups" element={<GroupsPage />} />
                <Route path="/groups/:groupId" element={<GroupDetailsPage />} />
                <Route path="/configs" element={<ConfigsPage />} />
                <Route path="/config-editor" element={<ConfigEditorPage />} />
                <Route path="/telemetry" element={<TelemetryPage />} />
                <Route path="/topology" element={<TopologyPage />} />
              </Route>
            </Routes>
          </Router>
        </ApiProvider>
      </SWRProvider>
    </ThemeProvider>
  );
}

export default App;