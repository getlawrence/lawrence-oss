import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { BarChart3, RefreshCw, Activity, Database, Zap } from 'lucide-react';

export default function TelemetryPage() {
    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold">Telemetry Explorer</h1>
                    <p className="text-gray-600">Explore metrics, logs, and traces from your agents</p>
                </div>
                <Button>
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Refresh
                </Button>
            </div>

            {/* Feature Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Metrics</CardTitle>
                        <BarChart3 className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">Coming Soon</div>
                        <p className="text-xs text-muted-foreground">
                            Query and visualize metrics from your agents
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Logs</CardTitle>
                        <Activity className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">Coming Soon</div>
                        <p className="text-xs text-muted-foreground">
                            Search and filter application logs
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Traces</CardTitle>
                        <Zap className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">Coming Soon</div>
                        <p className="text-xs text-muted-foreground">
                            Analyze distributed traces and performance
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* Status Card */}
            <Card>
                <CardHeader>
                    <CardTitle>Telemetry Status</CardTitle>
                    <CardDescription>
                        Current status of telemetry data collection and storage
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="space-y-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-2">
                                <Database className="h-5 w-5 text-blue-500" />
                                <span className="font-medium">DuckDB Storage</span>
                            </div>
                            <Badge variant="default" className="bg-green-100 text-green-800">
                                Connected
                            </Badge>
                        </div>
                        <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-2">
                                <Activity className="h-5 w-5 text-purple-500" />
                                <span className="font-medium">OTLP Ingestion</span>
                            </div>
                            <Badge variant="default" className="bg-green-100 text-green-800">
                                Active
                            </Badge>
                        </div>
                        <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-2">
                                <BarChart3 className="h-5 w-5 text-orange-500" />
                                <span className="font-medium">Query Engine</span>
                            </div>
                            <Badge variant="secondary">
                                Not Implemented
                            </Badge>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Coming Soon Message */}
            <Card>
                <CardHeader>
                    <CardTitle>Telemetry Features</CardTitle>
                    <CardDescription>
                        Advanced telemetry exploration capabilities
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="text-center py-8">
                        <BarChart3 className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                        <h3 className="text-lg font-medium text-gray-900 mb-2">Telemetry Explorer Coming Soon</h3>
                        <p className="text-gray-600 mb-4">
                            The telemetry query endpoints are not yet implemented in the current API.
                            This will include:
                        </p>
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-left max-w-2xl mx-auto">
                            <div className="space-y-2">
                                <h4 className="font-medium">Metrics</h4>
                                <ul className="text-sm text-gray-600 space-y-1">
                                    <li>• Time series queries</li>
                                    <li>• Aggregation functions</li>
                                    <li>• Custom dashboards</li>
                                </ul>
                            </div>
                            <div className="space-y-2">
                                <h4 className="font-medium">Logs</h4>
                                <ul className="text-sm text-gray-600 space-y-1">
                                    <li>• Full-text search</li>
                                    <li>• Log filtering</li>
                                    <li>• Real-time streaming</li>
                                </ul>
                            </div>
                            <div className="space-y-2">
                                <h4 className="font-medium">Traces</h4>
                                <ul className="text-sm text-gray-600 space-y-1">
                                    <li>• Trace visualization</li>
                                    <li>• Performance analysis</li>
                                    <li>• Service maps</li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
