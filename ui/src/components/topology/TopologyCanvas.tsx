import {
  ReactFlow,
  Background,
  useNodesState,
  useEdgesState,
  type Node,
} from "@xyflow/react";
import React from "react";

import "@xyflow/react/dist/style.css";
import { AgentNode } from "./AgentNode";
import { GroupNode } from "./GroupNode";

const nodeTypes = {
  agent: AgentNode,
  group: GroupNode,
};

interface TopologyCanvasProps {
  nodes: Node[];
  edges: any[];
  onNodeClick: (event: React.MouseEvent, node: Node) => void;
}

export function TopologyCanvas({
  nodes,
  edges,
  onNodeClick,
}: TopologyCanvasProps) {
  const [flowNodes, setNodes, onNodesChange] = useNodesState(nodes);
  const [flowEdges, setEdges, onEdgesChange] = useEdgesState(edges);

  // Update nodes and edges when props change
  React.useEffect(() => {
    setNodes(nodes);
    setEdges(edges);
  }, [nodes, edges, setNodes, setEdges]);

  return (
    <div className="flex-1 relative bg-gray-100">
      <ReactFlow
        nodes={flowNodes}
        edges={flowEdges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        nodeTypes={nodeTypes}
        fitView
        minZoom={0.2}
        maxZoom={2}
      >
        <Background />
      </ReactFlow>
    </div>
  );
}
