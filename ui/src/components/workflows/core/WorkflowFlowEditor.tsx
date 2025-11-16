import {
  ReactFlow,
  Background,
  addEdge,
  useNodesState,
  useEdgesState,
  type Connection,
  type NodeTypes,
  BackgroundVariant,
} from "@xyflow/react";
import { useCallback, useMemo, useState, useEffect, useRef } from "react";
import "@xyflow/react/dist/style.css";

import { ActionConfigDrawer } from "../config/ActionConfigDrawer";
import { extractWorkflowVariables } from "../utils/flow-utils";
import { BranchConfigDrawer } from "../config/BranchConfigDrawer";
import { ConditionConfigDrawer } from "../config/ConditionConfigDrawer";
import { DelayConfigDrawer } from "../config/DelayConfigDrawer";
import { ErrorHandlerConfigDrawer } from "../config/ErrorHandlerConfigDrawer";
import { GroupConfigDrawer } from "../config/GroupConfigDrawer";
import { LoopConfigDrawer } from "../config/LoopConfigDrawer";
import { NotificationConfigDrawer } from "../config/NotificationConfigDrawer";
import { ParallelConfigDrawer } from "../config/ParallelConfigDrawer";
import { SequentialConfigDrawer } from "../config/SequentialConfigDrawer";
import { TriggerConfigDrawer } from "../config/TriggerConfigDrawer";
import { VariableConfigDrawer } from "../config/VariableConfigDrawer";
import { ActionNode } from "../nodes/ActionNode";
import { BranchNode } from "../nodes/BranchNode";
import { ConditionNode } from "../nodes/ConditionNode";
import { DelayNode } from "../nodes/DelayNode";
import { ErrorHandlerNode } from "../nodes/ErrorHandlerNode";
import { GroupNode } from "../nodes/GroupNode";
import { LoopNode } from "../nodes/LoopNode";
import { NotificationNode } from "../nodes/NotificationNode";
import { ParallelNode } from "../nodes/ParallelNode";
import { SequentialNode } from "../nodes/SequentialNode";
import { TriggerNode } from "../nodes/TriggerNode";
import { VariableNode } from "../nodes/VariableNode";
import type {
  Workflow,
  FlowNode,
  NodeTemplate,
  TriggerNodeData,
  ConditionNodeData,
  LoopNodeData,
  DelayNodeData,
  NotificationNodeData,
  VariableNodeData,
  BranchNodeData,
  ParallelNodeData,
  SequentialNodeData,
  GroupNodeData,
  ErrorHandlerNodeData,
} from "../types/flow-types";

import { NodePalette } from "./NodePalette";

import { type WorkflowAction } from "@/api/workflows";

interface WorkflowFlowEditorProps {
  initialFlow?: Workflow;
  onSave?: (flow: Workflow) => void;
}

type DrawerState =
  | { type: "action"; nodeId: string; data: WorkflowAction }
  | { type: "trigger"; nodeId: string; data: TriggerNodeData }
  | { type: "condition"; nodeId: string; data: ConditionNodeData }
  | { type: "notification"; nodeId: string; data: NotificationNodeData }
  | { type: "parallel"; nodeId: string; data: ParallelNodeData }
  | { type: "sequential"; nodeId: string; data: SequentialNodeData }
  | { type: "loop"; nodeId: string; data: LoopNodeData }
  | { type: "delay"; nodeId: string; data: DelayNodeData }
  | { type: "group"; nodeId: string; data: GroupNodeData }
  | { type: "variable"; nodeId: string; data: VariableNodeData }
  | { type: "error-handler"; nodeId: string; data: ErrorHandlerNodeData }
  | { type: "branch"; nodeId: string; data: BranchNodeData }
  | null;

export function WorkflowFlowEditor({
  initialFlow,
  onSave,
}: WorkflowFlowEditorProps) {
  // Consolidated drawer state
  const [drawerState, setDrawerState] = useState<DrawerState>(null);

  // Ensure initial nodes have valid positions
  const initialNodes = useMemo(() => {
    if (!initialFlow?.nodes) return [];
    return initialFlow.nodes.map((node, index) => {
      if (!node.position || typeof node.position.x !== 'number' || typeof node.position.y !== 'number') {
        return {
          ...node,
          position: {
            x: 100 + (index % 3) * 350,
            y: 100 + Math.floor(index / 3) * 150,
          },
        };
      }
      return node;
    });
  }, [initialFlow?.nodes]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(
    initialFlow?.edges || [],
  );

  // Derive selected nodes from nodes state (no useEffect needed)
  const selectedNodes = useMemo(
    () => nodes.filter((n) => n.selected).map((n) => n.id),
    [nodes]
  );

  // Extract available variables from current flow
  const availableVariables = useMemo(() => {
    const currentFlow: Workflow = {
      nodes,
      edges,
      variables: initialFlow?.variables,
    };
    return extractWorkflowVariables(currentFlow);
  }, [nodes, edges, initialFlow?.variables]);

  const [paletteVisible, setPaletteVisible] = useState(true);

  // Use ref to track if initial flow has been set
  const initialFlowSetRef = useRef(false);
  // State counter to trigger sync effect (increment when we want to sync)
  const [syncTrigger, setSyncTrigger] = useState(0);
  // Refs to track latest nodes/edges for syncing (avoid dependency on state)
  const nodesRef = useRef(nodes);
  const edgesRef = useRef(edges);

  // Keep refs in sync with state
  useEffect(() => {
    nodesRef.current = nodes;
    edgesRef.current = edges;
  }, [nodes, edges]);

  // Sync initialFlow to nodes and edges when it changes
  // This is appropriate - syncing external data source to local state
  useEffect(() => {
    if (initialFlow && !initialFlowSetRef.current) {
      // Ensure all nodes have valid positions before setting them
      const nodesWithPositions = initialFlow.nodes.map((node, index) => {
        if (!node.position || typeof node.position.x !== 'number' || typeof node.position.y !== 'number') {
          return {
            ...node,
            position: {
              x: 100 + (index % 3) * 350,
              y: 100 + Math.floor(index / 3) * 150,
            },
          };
        }
        return node;
      });
      setNodes(nodesWithPositions);
      setEdges(initialFlow.edges);
      initialFlowSetRef.current = true;
    }
  }, [initialFlow, setNodes, setEdges]);

  // Sync flow state to parent only when syncTrigger changes (triggered by user actions)
  // Note: We don't depend on nodes/edges to avoid running on every ReactFlow update
  useEffect(() => {
    if (syncTrigger > 0 && onSave && nodesRef.current.length > 0) {
      onSave({ nodes: nodesRef.current, edges: edgesRef.current });
    }
  }, [syncTrigger, onSave]); // Only depend on syncTrigger and onSave, not nodes/edges

  // Node click handlers - simplified (all follow same pattern)
  const handleActionNodeClick = useCallback(
    (nodeId: string, action: WorkflowAction) => {
      setDrawerState({ type: "action", nodeId, data: action });
    },
    [],
  );

  const handleTriggerNodeClick = useCallback(
    (nodeId: string, data: TriggerNodeData) => {
      setDrawerState({ type: "trigger", nodeId, data });
    },
    [],
  );

  const handleConditionNodeClick = useCallback(
    (nodeId: string, data: ConditionNodeData) => {
      setDrawerState({ type: "condition", nodeId, data });
    },
    [],
  );

  const handleNotificationNodeClick = useCallback(
    (nodeId: string, data: NotificationNodeData) => {
      setDrawerState({ type: "notification", nodeId, data });
    },
    [],
  );

  const handleParallelNodeClick = useCallback(
    (nodeId: string, data: ParallelNodeData) => {
      setDrawerState({ type: "parallel", nodeId, data });
    },
    [],
  );

  const handleSequentialNodeClick = useCallback(
    (nodeId: string, data: SequentialNodeData) => {
      setDrawerState({ type: "sequential", nodeId, data });
    },
    [],
  );

  const handleLoopNodeClick = useCallback(
    (nodeId: string, data: LoopNodeData) => {
      setDrawerState({ type: "loop", nodeId, data });
    },
    [],
  );

  const handleDelayNodeClick = useCallback(
    (nodeId: string, data: DelayNodeData) => {
      setDrawerState({ type: "delay", nodeId, data });
    },
    [],
  );

  const handleGroupNodeClick = useCallback(
    (nodeId: string, data: GroupNodeData) => {
      setDrawerState({ type: "group", nodeId, data });
    },
    [],
  );

  const handleVariableNodeClick = useCallback(
    (nodeId: string, data: VariableNodeData) => {
      setDrawerState({ type: "variable", nodeId, data });
    },
    [],
  );

  const handleErrorHandlerNodeClick = useCallback(
    (nodeId: string, data: ErrorHandlerNodeData) => {
      setDrawerState({ type: "error-handler", nodeId, data });
    },
    [],
  );

  const handleBranchNodeClick = useCallback(
    (nodeId: string, data: BranchNodeData) => {
      setDrawerState({ type: "branch", nodeId, data });
    },
    [],
  );

  // Register all node types
  const nodeTypes: NodeTypes = useMemo(
    () => ({
      trigger: (props) => (
        <TriggerNode {...props} onNodeClick={handleTriggerNodeClick} />
      ),
      condition: (props) => (
        <ConditionNode {...props} onNodeClick={handleConditionNodeClick} />
      ),
      action: (props) => (
        <ActionNode {...props} onNodeClick={handleActionNodeClick} />
      ),
      parallel: (props) => (
        <ParallelNode {...props} onNodeClick={handleParallelNodeClick} />
      ),
      sequential: (props) => (
        <SequentialNode {...props} onNodeClick={handleSequentialNodeClick} />
      ),
      loop: (props) => (
        <LoopNode {...props} onNodeClick={handleLoopNodeClick} />
      ),
      delay: (props) => (
        <DelayNode {...props} onNodeClick={handleDelayNodeClick} />
      ),
      notification: (props) => (
        <NotificationNode
          {...props}
          onNodeClick={handleNotificationNodeClick}
        />
      ),
      group: (props) => (
        <GroupNode {...props} onNodeClick={handleGroupNodeClick} />
      ),
      variable: (props) => (
        <VariableNode {...props} onNodeClick={handleVariableNodeClick} />
      ),
      "error-handler": (props) => (
        <ErrorHandlerNode
          {...props}
          onNodeClick={handleErrorHandlerNodeClick}
        />
      ),
      branch: (props) => (
        <BranchNode {...props} onNodeClick={handleBranchNodeClick} />
      ),
    }),
    [
      handleActionNodeClick,
      handleTriggerNodeClick,
      handleConditionNodeClick,
      handleNotificationNodeClick,
      handleParallelNodeClick,
      handleSequentialNodeClick,
      handleLoopNodeClick,
      handleDelayNodeClick,
      handleGroupNodeClick,
      handleVariableNodeClick,
      handleErrorHandlerNodeClick,
      handleBranchNodeClick,
    ],
  );

  const onConnect = useCallback(
    (params: Connection) => {
      setEdges((eds) => addEdge(params, eds));
      setSyncTrigger((prev) => prev + 1); // Trigger sync
    },
    [setEdges],
  );

  // Add node from template
  const handleNodeSelect = useCallback(
    (template: NodeTemplate) => {
      // Calculate smart position for horizontal layout (left to right)
      const existingNodes = nodes.filter((n) => n.type === template.type);
      const baseX = 100 + Math.floor(nodes.length / 3) * 350;
      const baseY = 100 + (nodes.length % 3) * 150;

      const position =
        existingNodes.length === 0
          ? { x: baseX, y: baseY }
          : {
              x: existingNodes[existingNodes.length - 1].position.x + 350,
              y: existingNodes[existingNodes.length - 1].position.y,
            };

      const newNode: FlowNode = {
        id: `${template.type}-${Date.now()}`,
        type: template.type,
        position,
        data: template.defaultData,
      } as FlowNode;

      setNodes((nds) => [...nds, newNode]);
      setSyncTrigger((prev) => prev + 1); // Trigger sync
    },
    [nodes, setNodes],
  );

  // Delete selected nodes
  const deleteSelectedNodes = useCallback(() => {
    if (selectedNodes.length === 0) return;

    setNodes((nds) => nds.filter((node) => !selectedNodes.includes(node.id)));
    setEdges((eds) =>
      eds.filter(
        (edge) =>
          !selectedNodes.includes(edge.source) &&
          !selectedNodes.includes(edge.target),
      ),
    );
    setSyncTrigger((prev) => prev + 1); // Trigger sync
    // No need to manually clear selectedNodes - it's derived from nodes state
  }, [selectedNodes, setNodes, setEdges]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't intercept keys when typing in input fields
      const target = e.target as HTMLElement;
      const isInputField =
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable;

      if (
        (e.key === "Delete" || e.key === "Backspace") &&
        selectedNodes.length > 0 &&
        !isInputField
      ) {
        e.preventDefault();
        deleteSelectedNodes();
      }
      // Toggle palette with Cmd/Ctrl + P
      if (e.key === "p" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        setPaletteVisible((prev) => !prev);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [selectedNodes, deleteSelectedNodes]);

  // Shared helper for drawer save handlers
  const updateNodeData = useCallback(
    <T extends NonNullable<DrawerState>["type"]>(
      type: T,
      updatedData: Extract<NonNullable<DrawerState>, { type: T }>["data"],
    ) => {
      if (!drawerState || drawerState.type !== type) return;

      setNodes((nds) =>
        nds.map((node) => {
          if (node.id === drawerState.nodeId && node.type === type) {
            // Special handling for action node (nested data structure)
            if (type === "action") {
              return {
                ...node,
                data: {
                  ...node.data,
                  action: updatedData as WorkflowAction,
                },
              } as FlowNode;
            }
            // For all other types, replace data directly
            return { ...node, data: updatedData } as FlowNode;
          }
          return node;
        }),
      );
      setSyncTrigger((prev) => prev + 1);
      setDrawerState(null);
    },
    [drawerState, setNodes],
  );

  // Drawer save handlers - consolidated using shared helper
  const handleActionDrawerSave = useCallback(
    (updatedAction: WorkflowAction) => {
      updateNodeData("action", updatedAction);
    },
    [updateNodeData],
  );

  const handleTriggerDrawerSave = useCallback(
    (updatedData: TriggerNodeData) => {
      updateNodeData("trigger", updatedData);
    },
    [updateNodeData],
  );

  const handleConditionDrawerSave = useCallback(
    (updatedData: ConditionNodeData) => {
      updateNodeData("condition", updatedData);
    },
    [updateNodeData],
  );

  const handleNotificationDrawerSave = useCallback(
    (updatedData: NotificationNodeData) => {
      updateNodeData("notification", updatedData);
    },
    [updateNodeData],
  );

  const handleParallelDrawerSave = useCallback(
    (updatedData: ParallelNodeData) => {
      updateNodeData("parallel", updatedData);
    },
    [updateNodeData],
  );

  const handleSequentialDrawerSave = useCallback(
    (updatedData: SequentialNodeData) => {
      updateNodeData("sequential", updatedData);
    },
    [updateNodeData],
  );

  const handleLoopDrawerSave = useCallback(
    (updatedData: LoopNodeData) => {
      updateNodeData("loop", updatedData);
    },
    [updateNodeData],
  );

  const handleDelayDrawerSave = useCallback(
    (updatedData: DelayNodeData) => {
      updateNodeData("delay", updatedData);
    },
    [updateNodeData],
  );

  const handleGroupDrawerSave = useCallback(
    (updatedData: GroupNodeData) => {
      updateNodeData("group", updatedData);
    },
    [updateNodeData],
  );

  const handleVariableDrawerSave = useCallback(
    (updatedData: VariableNodeData) => {
      updateNodeData("variable", updatedData);
    },
    [updateNodeData],
  );

  const handleErrorHandlerDrawerSave = useCallback(
    (updatedData: ErrorHandlerNodeData) => {
      updateNodeData("error-handler", updatedData);
    },
    [updateNodeData],
  );

  const handleBranchDrawerSave = useCallback(
    (updatedData: BranchNodeData) => {
      updateNodeData("branch", updatedData);
    },
    [updateNodeData],
  );

  // Memoized drawer close handler - prevents re-renders
  const handleDrawerClose = useCallback((open: boolean) => {
    if (!open) setDrawerState(null);
  }, []);


  return (
    <div className="h-full w-full flex bg-background">
      {/* Node Palette Sidebar */}
      {paletteVisible && (
        <div className="w-72 flex-shrink-0">
          <NodePalette onNodeSelect={handleNodeSelect} />
        </div>
      )}

      {/* Main Flow Editor */}
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          nodeTypes={nodeTypes}
          fitView
          deleteKeyCode={["Delete", "Backspace"]}
          multiSelectionKeyCode={["Meta", "Control"]}
          defaultEdgeOptions={{
            style: { strokeWidth: 2 },
            type: "smoothstep",
            animated: false,
          }}
          proOptions={{ hideAttribution: true }}
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={20}
            size={1.5}
            color="hsl(var(--muted-foreground))"
            className="opacity-20"
          />
        </ReactFlow>
      </div>

      {/* Configuration Drawers */}
      <ActionConfigDrawer
        open={drawerState?.type === "action"}
        onOpenChange={handleDrawerClose}
        action={drawerState?.type === "action" ? drawerState.data : null}
        onSave={handleActionDrawerSave}
        availableVariables={availableVariables}
      />

      <TriggerConfigDrawer
        open={drawerState?.type === "trigger"}
        onOpenChange={handleDrawerClose}
        nodeData={
          drawerState?.type === "trigger" ? drawerState.data : null
        }
        onSave={handleTriggerDrawerSave}
      />

      <ConditionConfigDrawer
        open={drawerState?.type === "condition"}
        onOpenChange={handleDrawerClose}
        nodeData={
          drawerState?.type === "condition" ? drawerState.data : null
        }
        onSave={handleConditionDrawerSave}
      />

      <NotificationConfigDrawer
        open={drawerState?.type === "notification"}
        onOpenChange={handleDrawerClose}
        nodeData={
          drawerState?.type === "notification" ? drawerState.data : null
        }
        onSave={handleNotificationDrawerSave}
      />

      <ParallelConfigDrawer
        open={drawerState?.type === "parallel"}
        onOpenChange={handleDrawerClose}
        nodeData={
          drawerState?.type === "parallel" ? drawerState.data : null
        }
        onSave={handleParallelDrawerSave}
      />

      <SequentialConfigDrawer
        open={drawerState?.type === "sequential"}
        onOpenChange={handleDrawerClose}
        nodeData={
          drawerState?.type === "sequential" ? drawerState.data : null
        }
        onSave={handleSequentialDrawerSave}
      />

      <LoopConfigDrawer
        open={drawerState?.type === "loop"}
        onOpenChange={handleDrawerClose}
        nodeData={drawerState?.type === "loop" ? drawerState.data : null}
        onSave={handleLoopDrawerSave}
      />

      <DelayConfigDrawer
        open={drawerState?.type === "delay"}
        onOpenChange={handleDrawerClose}
        nodeData={drawerState?.type === "delay" ? drawerState.data : null}
        onSave={handleDelayDrawerSave}
      />

      <GroupConfigDrawer
        open={drawerState?.type === "group"}
        onOpenChange={handleDrawerClose}
        nodeData={drawerState?.type === "group" ? drawerState.data : null}
        onSave={handleGroupDrawerSave}
      />

      <VariableConfigDrawer
        open={drawerState?.type === "variable"}
        onOpenChange={handleDrawerClose}
        nodeData={
          drawerState?.type === "variable" ? drawerState.data : null
        }
        onSave={handleVariableDrawerSave}
      />

      <ErrorHandlerConfigDrawer
        open={drawerState?.type === "error-handler"}
        onOpenChange={handleDrawerClose}
        nodeData={
          drawerState?.type === "error-handler" ? drawerState.data : null
        }
        onSave={handleErrorHandlerDrawerSave}
      />

      <BranchConfigDrawer
        open={drawerState?.type === "branch"}
        onOpenChange={handleDrawerClose}
        nodeData={drawerState?.type === "branch" ? drawerState.data : null}
        onSave={handleBranchDrawerSave}
      />
    </div>
  );
}
