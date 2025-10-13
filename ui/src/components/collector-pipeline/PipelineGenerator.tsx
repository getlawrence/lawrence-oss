import type { Node, Edge } from "@xyflow/react";
import * as yaml from "js-yaml";

import type { ComponentMetrics } from "@/api/collector-metrics";

// Helper function to find metrics for a specific component
function findComponentMetrics(
  metrics: ComponentMetrics[],
  componentType: string,
  componentName: string,
  pipelineType: string,
): ComponentMetrics | undefined {
  return metrics.find(
    (m) =>
      m.component_type === componentType &&
      m.component_name === componentName &&
      m.pipeline_type === pipelineType,
  );
}

// Generate nodes and edges for the pipeline based on actual agent configuration
export function generatePipelineNodes(
  effectiveConfig: string,
  metrics: ComponentMetrics[] = [],
): {
  nodes: Node[];
  edges: Edge[];
} {
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  let yOffset = 0;

  // Check if effective config exists
  if (!effectiveConfig) {
    return {
      nodes: [
        {
          id: "no-config",
          type: "default",
          position: { x: 300, y: 200 },
          data: {
            label: "No configuration available",
          },
        },
      ],
      edges: [],
    };
  }

  // Parse the YAML configuration
  let parsedConfig;
  try {
    parsedConfig = yaml.load(effectiveConfig) as any;
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : "Error parsing configuration";
    return {
      nodes: [
        {
          id: "parse-error",
          type: "default",
          position: { x: 300, y: 200 },
          data: {
            label: `YAML Parse Error: ${errorMessage}`,
          },
        },
      ],
      edges: [],
    };
  }

  if (!parsedConfig.service || !parsedConfig.service.pipelines) {
    return {
      nodes: [
        {
          id: "no-service",
          type: "default",
          position: { x: 300, y: 200 },
          data: {
            label: "No service configuration found",
          },
        },
      ],
      edges: [],
    };
  }

  const { pipelines } = parsedConfig.service;

  // Check if there are any pipelines configured
  if (!pipelines || Object.keys(pipelines).length === 0) {
    return {
      nodes: [
        {
          id: "no-pipelines",
          type: "default",
          position: { x: 300, y: 200 },
          data: {
            label: "No pipelines configured",
          },
        },
      ],
      edges: [],
    };
  }

  // Create pipeline nodes based on actual configuration
  Object.entries(pipelines).forEach(
    ([pipelineName, pipelineConfig]: [string, any]) => {
      const pipelineType = pipelineName.toLowerCase();

      // Determine pipeline display name
      const displayName =
        pipelineName === "traces" ||
        pipelineName === "metrics" ||
        pipelineName === "logs"
          ? pipelineName.toUpperCase()
          : `${pipelineType.toUpperCase()} (${pipelineName})`;

      // Get the actual components used in THIS pipeline from the service.pipelines section
      const pipelineReceivers = pipelineConfig?.receivers || [];
      const pipelineProcessors = pipelineConfig?.processors || [];
      const pipelineExporters = pipelineConfig?.exporters || [];

      // Calculate aggregate metrics for this pipeline
      const pipelineMetrics = metrics.filter(
        (m) => m.pipeline_type === pipelineType,
      );
      const totalReceived = pipelineMetrics.reduce(
        (sum, m) => sum + (m.received || 0),
        0,
      );
      const totalErrors = pipelineMetrics.reduce((sum, m) => sum + m.errors, 0);

      // Create section container node (styled)
      const sectionNode: Node = {
        id: `section-${pipelineName}`,
        type: "section",
        position: { x: 50, y: yOffset },
        data: {
          type: pipelineType as "traces" | "metrics" | "logs",
          label: displayName,
          width: 850,
          height: 320,
          metrics: {
            received: totalReceived,
            errors: totalErrors,
          },
        },
        selectable: false,
        draggable: false,
      };
      nodes.push(sectionNode);

      // Calculate positions for proper pipeline flow: Fan-in → Chain → Fan-out
      const sectionHeight = 320;
      const centerY = yOffset + sectionHeight / 2;

      // Fan-in: Receivers positioned to converge toward center
      const receiverCount = pipelineReceivers.length;
      const receiverSpacing = Math.min(
        80,
        (sectionHeight - 100) / Math.max(1, receiverCount - 1),
      );
      const receiverStartY =
        centerY - ((receiverCount - 1) * receiverSpacing) / 2;

      pipelineReceivers.forEach((receiver: string, index: number) => {
        // Find metrics for this receiver
        const receiverMetrics = findComponentMetrics(
          metrics,
          "receiver",
          receiver,
          pipelineType,
        );

        const receiverNode: Node = {
          id: `receiver-${pipelineType}-${receiver}`,
          type: "receiver",
          position: {
            x: 100,
            y: receiverStartY + index * receiverSpacing,
          },
          data: {
            label: receiver,
            pipelineType: pipelineType,
            metrics: {
              received:
                receiverMetrics?.received || receiverMetrics?.accepted || 0,
            },
          },
        };
        nodes.push(receiverNode);
      });

      // Chain: Processors positioned vertically in the center
      const processorSpacing = Math.min(
        100,
        (sectionHeight - 100) / Math.max(1, pipelineProcessors.length - 1),
      );
      const processorStartY =
        centerY - ((pipelineProcessors.length - 1) * processorSpacing) / 2;

      pipelineProcessors.forEach((processor: string, index: number) => {
        // Find metrics for this processor
        const processorMetrics = findComponentMetrics(
          metrics,
          "processor",
          processor,
          pipelineType,
        );

        const processorNode: Node = {
          id: `processor-${pipelineType}-${processor}`,
          type: "processor",
          position: {
            x: 350,
            y: processorStartY + index * processorSpacing,
          },
          data: {
            label: processor,
            pipelineType: pipelineType,
            metrics: {
              processed:
                processorMetrics?.accepted || processorMetrics?.received || 0,
              batches: 0, // batches info not available in current metrics
            },
          },
        };
        nodes.push(processorNode);
      });

      // Fan-out: Exporters positioned to diverge from center
      const exporterCount = pipelineExporters.length;
      const exporterSpacing = Math.min(
        80,
        (sectionHeight - 100) / Math.max(1, exporterCount - 1),
      );
      const exporterStartY =
        centerY - ((exporterCount - 1) * exporterSpacing) / 2;

      pipelineExporters.forEach((exporter: string, index: number) => {
        // Find metrics for this exporter
        const exporterMetrics = findComponentMetrics(
          metrics,
          "exporter",
          exporter,
          pipelineType,
        );

        const exporterNode: Node = {
          id: `exporter-${pipelineType}-${exporter}`,
          type: "exporter",
          position: {
            x: 600,
            y: exporterStartY + index * exporterSpacing,
          },
          data: {
            label: exporter,
            pipelineType: pipelineType,
            metrics: {
              exported: exporterMetrics?.sent || 0,
            },
          },
        };
        nodes.push(exporterNode);
      });

      // Create connections following OTEL pipeline logic:
      // Fan-in: Receivers → Processors (or Exporters if no processors)
      // Chain: Processors → Processors (in sequence)
      // Fan-out: Last processor → Exporters (or Receivers → Exporters if no processors)

      if (pipelineProcessors.length > 0) {
        // Fan-in: Connect all receivers to first processor (converging pattern)
        pipelineReceivers.forEach((receiver: string) => {
          edges.push({
            id: `edge-${pipelineType}-${receiver}-to-${pipelineProcessors[0]}`,
            source: `receiver-${pipelineType}-${receiver}`,
            target: `processor-${pipelineType}-${pipelineProcessors[0]}`,
            type: "smoothstep",
            animated: true,
            style: {
              stroke: "#ff9800",
              strokeWidth: 3,
              zIndex: 1000,
              // Add slight curve to show fan-in pattern
              strokeDasharray: "0",
            },
          });
        });

        // Chain: Connect processors in sequence (vertical chain)
        for (let i = 0; i < pipelineProcessors.length - 1; i++) {
          const currentProcessor = pipelineProcessors[i];
          const nextProcessor = pipelineProcessors[i + 1];
          if (currentProcessor && nextProcessor) {
            edges.push({
              id: `edge-${pipelineType}-${currentProcessor}-to-${nextProcessor}`,
              source: `processor-${pipelineType}-${currentProcessor}`,
              target: `processor-${pipelineType}-${nextProcessor}`,
              type: "smoothstep",
              animated: true,
              style: {
                stroke: "#ff9800",
                strokeWidth: 3,
                zIndex: 1000,
              },
            });
          }
        }

        // Fan-out: Connect last processor to all exporters (diverging pattern)
        const lastProcessor = pipelineProcessors[pipelineProcessors.length - 1];
        pipelineExporters.forEach((exporter: string) => {
          edges.push({
            id: `edge-${pipelineType}-${lastProcessor}-to-${exporter}`,
            source: `processor-${pipelineType}-${lastProcessor}`,
            target: `exporter-${pipelineType}-${exporter}`,
            type: "smoothstep",
            animated: true,
            style: {
              stroke: "#ff9800",
              strokeWidth: 3,
              zIndex: 1000,
              // Add slight curve to show fan-out pattern
              strokeDasharray: "0",
            },
          });
        });
      } else {
        // No processors: Direct fan-in to fan-out (receivers → exporters)
        pipelineReceivers.forEach((receiver: string) => {
          pipelineExporters.forEach((exporter: string) => {
            edges.push({
              id: `edge-${pipelineType}-${receiver}-to-${exporter}`,
              source: `receiver-${pipelineType}-${receiver}`,
              target: `exporter-${pipelineType}-${exporter}`,
              type: "smoothstep",
              animated: true,
              style: {
                stroke: "#ff9800",
                strokeWidth: 3,
                zIndex: 1000,
              },
            });
          });
        });
      }

      // Move yOffset for next pipeline (based on section height + gap)
      yOffset += 320 + 40; // Match new height + reduced gap
    },
  );

  return { nodes, edges };
}
