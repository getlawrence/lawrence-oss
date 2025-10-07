import { LawrenceQLInterface } from '@/components/lawrence-ql/LawrenceQLInterface';

export default function TelemetryPage() {
    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold">Telemetry Explorer</h1>
                    <p className="text-gray-600">Query and explore metrics, logs, and traces using Lawrence QL</p>
                </div>
            </div>

            {/* Lawrence QL Interface */}
            <div className="min-h-[600px]">
                <LawrenceQLInterface />
            </div>
        </div>
    );
}
