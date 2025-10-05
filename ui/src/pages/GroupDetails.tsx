import { useState, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import useSWR from 'swr';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  RefreshCw,
  Users,
  Server,
  CheckCircle,
  XCircle,
  AlertCircle,
  BarChart3,
  Settings,
  FileText,
} from 'lucide-react';
import { getGroupTopology } from '@/api/topology';
import { getGroup, assignConfigToGroup, getGroupConfig, getGroupAgents } from '@/api/groups';
import { getConfigs } from '@/api/configs';
import { queryMetrics, type MetricData } from '@/api/telemetry';
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';

export default function GroupDetailsPage() {
  const { groupId } = useParams<{ groupId: string }>();
  const [refreshing, setRefreshing] = useState(false);
  const [showConfigDialog, setShowConfigDialog] = useState(false);
  const [selectedConfigId, setSelectedConfigId] = useState('');
  const [metricsData, setMetricsData] = useState<MetricData[]>([]);

  const { data: groupData, error, mutate } = useSWR(
    groupId ? `group-${groupId}` : null,
    () => groupId ? getGroup(groupId) : null
  );

  useSWR(
    groupId ? `group-topology-${groupId}` : null,
    () => groupId ? getGroupTopology(groupId) : null
  );

  const { data: agentsData } = useSWR(
    groupId ? `group-agents-${groupId}` : null,
    () => groupId ? getGroupAgents(groupId) : null
  );

  const { data: configData } = useSWR(
    groupId ? `group-config-${groupId}` : null,
    () => groupId ? getGroupConfig(groupId).catch(() => null) : null
  );

  const { data: allConfigsData } = useSWR('all-configs', getConfigs);

  // Fetch group metrics
  useSWR(
    groupId ? `group-metrics-${groupId}` : null,
    async () => {
      if (!groupId) return null;
      const endTime = new Date();
      const startTime = new Date(endTime.getTime() - 4 * 60 * 60 * 1000); // Last 4 hours

      const result = await queryMetrics({
        group_id: groupId,
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
        limit: 500,
      });
      setMetricsData(result.metrics);
      return result;
    },
    { refreshInterval: 60000 }
  );

  const handleRefresh = async () => {
    setRefreshing(true);
    await mutate();
    setRefreshing(false);
  };

  const handleAssignConfig = async () => {
    if (!groupId || !selectedConfigId) return;

    try {
      await assignConfigToGroup(groupId, { config_id: selectedConfigId });
      setShowConfigDialog(false);
      mutate();
      alert('Configuration assigned successfully!');
    } catch (error) {
      console.error('Failed to assign config:', error);
      alert('Failed to assign configuration');
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'offline':
        return <XCircle className="h-4 w-4 text-gray-500" />;
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Server className="h-4 w-4 text-gray-400" />;
    }
  };

  // Prepare chart data
  const chartData = useMemo(() => {
    if (!metricsData.length) return [];

    // Group by 15-minute intervals
    const grouped = metricsData.reduce((acc, metric) => {
      const time = new Date(metric.timestamp);
      const interval = new Date(Math.floor(time.getTime() / (15 * 60 * 1000)) * (15 * 60 * 1000));
      const key = interval.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

      if (!acc[key]) {
        acc[key] = { time: key, count: 0, totalValue: 0 };
      }
      acc[key].count += 1;
      acc[key].totalValue += metric.value;
      return acc;
    }, {} as Record<string, { time: string; count: number; totalValue: number }>);

    return Object.values(grouped).sort((a, b) => a.time.localeCompare(b.time));
  }, [metricsData]);

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">Error Loading Group</h1>
          <p className="text-gray-600">{error.message}</p>
          <Button onClick={handleRefresh} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  if (!groupData) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading group details...</p>
        </div>
      </div>
    );
  }

  const agents = agentsData?.agents || [];
  const agentStats = {
    total: agents.length,
    online: agents.filter(a => a.status === 'online').length,
    offline: agents.filter(a => a.status === 'offline').length,
    error: agents.filter(a => a.status === 'error').length,
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <Users className="h-8 w-8 text-purple-500" />
            {groupData.name}
          </h1>
          <p className="text-gray-600">Group ID: {groupData.id}</p>
        </div>
        <div className="flex gap-2">
          <Dialog open={showConfigDialog} onOpenChange={setShowConfigDialog}>
            <DialogTrigger asChild>
              <Button variant="outline">
                <Settings className="h-4 w-4 mr-2" />
                Assign Config
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Assign Configuration</DialogTitle>
                <DialogDescription>
                  Select a configuration to assign to this group
                </DialogDescription>
              </DialogHeader>
              <div className="py-4">
                <Select value={selectedConfigId} onValueChange={setSelectedConfigId}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a configuration" />
                  </SelectTrigger>
                  <SelectContent>
                    {allConfigsData?.configs.map((config) => (
                      <SelectItem key={config.id} value={config.id}>
                        Version {config.version} - {new Date(config.created_at).toLocaleDateString()}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setShowConfigDialog(false)}>
                  Cancel
                </Button>
                <Button onClick={handleAssignConfig} disabled={!selectedConfigId}>
                  Assign
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
          <Button onClick={handleRefresh} disabled={refreshing}>
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Agents</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{agentStats.total}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Online</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{agentStats.online}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Offline</CardTitle>
            <XCircle className="h-4 w-4 text-gray-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-gray-600">{agentStats.offline}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Metrics (4h)</CardTitle>
            <BarChart3 className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">{metricsData.length}</div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="agents" className="space-y-4">
        <TabsList>
          <TabsTrigger value="agents">Agents</TabsTrigger>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
          <TabsTrigger value="config">Configuration</TabsTrigger>
        </TabsList>

        <TabsContent value="agents" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Agents in Group</CardTitle>
              <CardDescription>{agents.length} agents currently in this group</CardDescription>
            </CardHeader>
            <CardContent>
              {agents.length === 0 ? (
                <div className="text-center py-8">
                  <Server className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600">No agents in this group</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Status</TableHead>
                      <TableHead>Name</TableHead>
                      <TableHead>Version</TableHead>
                      <TableHead>Last Seen</TableHead>
                      <TableHead>Labels</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {agents.map((agent) => (
                      <TableRow key={agent.id}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getStatusIcon(agent.status)}
                          </div>
                        </TableCell>
                        <TableCell className="font-medium">{agent.name}</TableCell>
                        <TableCell>{agent.version}</TableCell>
                        <TableCell>{new Date(agent.last_seen).toLocaleString()}</TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-1">
                            {Object.entries(agent.labels).slice(0, 2).map(([key, value]) => (
                              <Badge key={key} variant="outline" className="text-xs">
                                {key}={value}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="metrics" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Group Metrics Timeline</CardTitle>
              <CardDescription>Aggregated metrics from all agents in the last 4 hours</CardDescription>
            </CardHeader>
            <CardContent>
              {chartData.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" style={{ fontSize: '11px' }} />
                    <YAxis style={{ fontSize: '11px' }} />
                    <Tooltip />
                    <Line type="monotone" dataKey="count" stroke="#3b82f6" strokeWidth={2} name="Metric Count" />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="text-center py-12 text-gray-500">
                  No metrics data available for this group
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="config" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="h-5 w-5" />
                Active Configuration
              </CardTitle>
              <CardDescription>
                Configuration currently assigned to this group
              </CardDescription>
            </CardHeader>
            <CardContent>
              {configData ? (
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Config ID:</span>
                    <span className="text-sm font-mono">{configData.id}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Version:</span>
                    <span className="text-sm">{configData.version}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Created:</span>
                    <span className="text-sm">{new Date(configData.created_at).toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Hash:</span>
                    <span className="text-sm font-mono text-xs">{configData.config_hash.substring(0, 16)}...</span>
                  </div>
                </div>
              ) : (
                <div className="text-center py-8">
                  <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600">No configuration assigned to this group</p>
                  <Button onClick={() => setShowConfigDialog(true)} className="mt-4">
                    <Settings className="h-4 w-4 mr-2" />
                    Assign Config
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
