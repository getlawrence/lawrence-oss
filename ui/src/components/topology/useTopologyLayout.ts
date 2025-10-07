import { useMemo } from 'react';
import { type Node, type Edge, Position, MarkerType } from '@xyflow/react';
import type { TopologyData } from './types';

export function useTopologyLayout(topologyData: TopologyData | undefined, topologyLevel: 'instance' | 'group') {
  const nodes: Node[] = useMemo(() => {
    if (!topologyData?.nodes || !Array.isArray(topologyData.nodes) || topologyData.nodes.length === 0) return [];

    // Calculate layout positions
    const dataNodes = topologyData.nodes;
    const agentNodes = dataNodes.filter(n => n.type === 'agent');
    const groupNodes = dataNodes.filter(n => n.type === 'group');

    const flowNodes: Node[] = [];

    // Filter nodes based on topology level
    if (topologyLevel === 'group') {
      // Only show group nodes
      groupNodes.forEach((node, idx) => {
        flowNodes.push({
          id: node.id,
          type: 'group',
          data: {
            name: node.name,
            status: node.status,
            agent_count: topologyData.groups?.find(g => g.id === node.id.replace('group-', ''))?.agent_count || 0,
          },
          position: { x: idx * 300 + 50, y: 150 },
          sourcePosition: Position.Bottom,
          targetPosition: Position.Top,
        });
      });
      return flowNodes;
    }

    // Position groups in top row
    groupNodes.forEach((node, idx) => {
      flowNodes.push({
        id: node.id,
        type: 'group',
        data: {
          name: node.name,
          status: node.status,
          agent_count: topologyData.groups?.find(g => g.id === node.id.replace('group-', ''))?.agent_count || 0,
        },
        position: { x: idx * 300 + 50, y: 50 },
        sourcePosition: Position.Bottom,
        targetPosition: Position.Top,
      });
    });

    // Position agents in rows below groups
    const agentsPerRow = 4;
    agentNodes.forEach((node, idx) => {
      const row = Math.floor(idx / agentsPerRow);
      const col = idx % agentsPerRow;

      flowNodes.push({
        id: node.id,
        type: 'agent',
        data: {
          name: node.name,
          status: node.status,
          group_name: node.group_name,
          metrics: node.metrics,
          labels: node.labels,
        },
        position: { x: col * 250 + 50, y: row * 200 + 300 },
        sourcePosition: Position.Top,
        targetPosition: Position.Bottom,
      });
    });

    return flowNodes;
  }, [topologyData, topologyLevel]);

  const edges: Edge[] = useMemo(() => {
    if (!topologyData?.edges || !Array.isArray(topologyData.edges)) return [];

    // Filter edges based on topology level
    const filteredEdges = topologyData.edges.filter(edge => {
      if (topologyLevel === 'group') {
        // Only show edges between groups
        return edge.source.startsWith('group-') && edge.target.startsWith('group-');
      }
      // Instance level: show all edges
      return true;
    });

    return filteredEdges.map((edge, idx) => ({
      id: `edge-${idx}`,
      source: edge.source,
      target: edge.target,
      type: 'smoothstep',
      animated: false,
      label: edge.label,
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: '#9333ea',
      },
      style: {
        stroke: '#9333ea',
        strokeWidth: 2,
      },
    }));
  }, [topologyData, topologyLevel]);

  return { nodes, edges };
}
