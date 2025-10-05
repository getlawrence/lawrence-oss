import { useState } from 'react';
import useSWR from 'swr';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { getAgents, getAgentStats } from '@/api/agents';
import { RefreshCw, Server, AlertCircle, CheckCircle, XCircle } from 'lucide-react';

export default function InventoryPage() {
    const [refreshing, setRefreshing] = useState(false);

    const { data: agentsData, error: agentsError, mutate: mutateAgents } = useSWR(
        'agents',
        getAgents,
        { refreshInterval: 30000 }
    );

    const { data: statsData, error: statsError } = useSWR(
        'agent-stats',
        getAgentStats,
        { refreshInterval: 30000 }
    );

    const handleRefresh = async () => {
        setRefreshing(true);
        await mutateAgents();
        setRefreshing(false);
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
                return <Server className="h-4 w-4 text-gray-500" />;
        }
    };

    const getStatusBadge = (status: string) => {
        switch (status) {
            case 'online':
                return <Badge variant="default" className="bg-green-100 text-green-800">Online</Badge>;
            case 'offline':
                return <Badge variant="secondary">Offline</Badge>;
            case 'error':
                return <Badge variant="destructive">Error</Badge>;
            default:
                return <Badge variant="outline">Unknown</Badge>;
        }
    };

    if (agentsError || statsError) {
        return (
            <div className="container mx-auto p-6">
                <div className="text-center">
                    <h1 className="text-2xl font-bold text-red-600 mb-4">Error Loading Data</h1>
                    <p className="text-gray-600">
                        {agentsError?.message || statsError?.message || 'Failed to load agent data'}
                    </p>
                    <Button onClick={handleRefresh} className="mt-4">
                        <RefreshCw className="h-4 w-4 mr-2" />
                        Retry
                    </Button>
                </div>
            </div>
        );
    }

    const agents = agentsData?.agents ? Object.values(agentsData.agents) : [];
    const stats = statsData || { totalAgents: 0, onlineAgents: 0, offlineAgents: 0, errorAgents: 0, groupsCount: 0 };

    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold">Agent Inventory</h1>
                    <p className="text-gray-600">Manage and monitor your OpenTelemetry agents</p>
                </div>
                <Button onClick={handleRefresh} disabled={refreshing}>
                    <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
                    Refresh
                </Button>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Total Agents</CardTitle>
                        <Server className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.totalAgents}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Online</CardTitle>
                        <CheckCircle className="h-4 w-4 text-green-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-green-600">{stats.onlineAgents}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Offline</CardTitle>
                        <XCircle className="h-4 w-4 text-gray-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-gray-600">{stats.offlineAgents}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Errors</CardTitle>
                        <AlertCircle className="h-4 w-4 text-red-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-red-600">{stats.errorAgents}</div>
                    </CardContent>
                </Card>
            </div>

            {/* Agents Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Agents ({agents.length})</CardTitle>
                    <CardDescription>
                        All registered OpenTelemetry agents and their current status
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {agents.length === 0 ? (
                        <div className="text-center py-8">
                            <Server className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                            <h3 className="text-lg font-medium text-gray-900 mb-2">No Agents Found</h3>
                            <p className="text-gray-600">
                                No agents are currently registered. Agents will appear here once they connect.
                            </p>
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Version</TableHead>
                                    <TableHead>Group</TableHead>
                                    <TableHead>Last Seen</TableHead>
                                    <TableHead>Labels</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {agents.map((agent) => (
                                    <TableRow key={agent.id}>
                                        <TableCell>
                                            <div className="flex items-center space-x-2">
                                                {getStatusIcon(agent.status)}
                                                {getStatusBadge(agent.status)}
                                            </div>
                                        </TableCell>
                                        <TableCell className="font-medium">{agent.name}</TableCell>
                                        <TableCell>{agent.version}</TableCell>
                                        <TableCell>{agent.group_id || 'No Group'}</TableCell>
                                        <TableCell>
                                            {new Date(agent.last_seen).toLocaleString()}
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {Object.entries(agent.labels).slice(0, 2).map(([key, value]) => (
                                                    <Badge key={key} variant="outline" className="text-xs">
                                                        {key}={value}
                                                    </Badge>
                                                ))}
                                                {Object.keys(agent.labels).length > 2 && (
                                                    <Badge variant="outline" className="text-xs">
                                                        +{Object.keys(agent.labels).length - 2} more
                                                    </Badge>
                                                )}
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
