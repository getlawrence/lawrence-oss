import type { NodeTemplate, NodeCategory } from "../types/flow-types";

// Node templates for the palette
export const nodeTemplates: NodeTemplate[] = [
  // Trigger nodes
  {
    id: "trigger-manual",
    type: "trigger",
    name: "Manual Trigger",
    description: "Execute workflow manually",
    icon: "Zap",
    defaultData: {
      label: "Manual Trigger",
      triggerType: "manual",
    },
    category: "triggers",
    color: "blue",
    ports: { inputs: 0, outputs: 1 },
  },
  {
    id: "trigger-schedule",
    type: "trigger",
    name: "Schedule Trigger",
    description: "Execute workflow on a schedule",
    icon: "Calendar",
    defaultData: {
      label: "Schedule Trigger",
      triggerType: "schedule",
      cronExpression: "0 * * * *",
      timezone: "UTC",
    },
    category: "triggers",
    color: "blue",
    ports: { inputs: 0, outputs: 1 },
  },
  {
    id: "trigger-webhook",
    type: "trigger",
    name: "Webhook Trigger",
    description: "Execute workflow via webhook",
    icon: "Webhook",
    defaultData: {
      label: "Webhook Trigger",
      triggerType: "webhook",
    },
    category: "triggers",
    color: "blue",
    ports: { inputs: 0, outputs: 1 },
  },
  {
    id: "trigger-telemetry",
    type: "trigger",
    name: "Telemetry Trigger",
    description: "Execute workflow based on collector telemetry",
    icon: "Activity",
    defaultData: {
      label: "Telemetry Trigger",
      triggerType: "telemetry",
      telemetryConfig: {
        type: "log",
      },
    },
    category: "triggers",
    color: "blue",
    ports: { inputs: 0, outputs: 1 },
  },

  // Logic nodes
  {
    id: "condition",
    type: "condition",
    name: "Condition",
    description: "Filter execution based on conditions",
    icon: "Filter",
    defaultData: {
      label: "Condition",
      conditions: [],
      logic: "and",
    },
    category: "logic",
    color: "amber",
    ports: { inputs: 1, outputs: 1 },
  },
  {
    id: "branch",
    type: "branch",
    name: "Branch",
    description: "Multi-way branching based on conditions",
    icon: "GitBranch",
    defaultData: {
      label: "Branch",
      branches: [
        { name: "Branch 1", condition: [], isDefault: false },
        { name: "Default", condition: [], isDefault: true },
      ],
    },
    category: "logic",
    color: "amber",
    ports: { inputs: 1, outputs: -1 }, // Dynamic outputs
  },

  // Action nodes
  {
    id: "action-config",
    type: "action",
    name: "Config Update",
    description: "Update agent or group configuration",
    icon: "Settings",
    defaultData: {
      label: "Config Update",
      action: {
        type: "config_update",
        target_type: "group",
        target_id: "",
        config_update: {
          operation: "replace",
          template: "",
        },
      },
      continueOnError: false,
    },
    category: "actions",
    color: "green",
    ports: { inputs: 1, outputs: 1 },
  },
  {
    id: "notification",
    type: "notification",
    name: "Notification",
    description: "Send notifications",
    icon: "Bell",
    defaultData: {
      label: "Send Notification",
      channel: "log",
      message: "",
      severity: "info",
    },
    category: "actions",
    color: "purple",
    ports: { inputs: 1, outputs: 1 },
  },

  // Control flow nodes
  {
    id: "parallel",
    type: "parallel",
    name: "Parallel",
    description: "Execute multiple actions in parallel",
    icon: "Layers",
    defaultData: {
      label: "Parallel Execution",
      waitForAll: true,
      timeout: 300,
    },
    category: "control",
    color: "indigo",
    ports: { inputs: 1, outputs: -1 }, // Dynamic outputs
  },
  {
    id: "sequential",
    type: "sequential",
    name: "Sequential",
    description: "Execute actions in sequence",
    icon: "ArrowRight",
    defaultData: {
      label: "Sequential Execution",
      delayBetween: 0,
    },
    category: "control",
    color: "indigo",
    ports: { inputs: 1, outputs: -1 }, // Dynamic outputs
  },
  {
    id: "loop",
    type: "loop",
    name: "Loop",
    description: "Loop through agents or groups",
    icon: "Repeat",
    defaultData: {
      label: "Loop",
      loopType: "agents",
      maxIterations: 100,
      parallelExecution: false,
    },
    category: "control",
    color: "indigo",
    ports: { inputs: 1, outputs: 1 },
  },
  {
    id: "delay",
    type: "delay",
    name: "Delay",
    description: "Add time delay",
    icon: "Clock",
    defaultData: {
      label: "Delay",
      duration: 5,
      unit: "seconds",
    },
    category: "control",
    color: "slate",
    ports: { inputs: 1, outputs: 1 },
  },

  // Advanced nodes
  {
    id: "variable",
    type: "variable",
    name: "Variable",
    description: "Set or get variables",
    icon: "Variable",
    defaultData: {
      label: "Variable",
      operation: "set",
      variableName: "",
      value: "",
    },
    category: "advanced",
    color: "cyan",
    ports: { inputs: 1, outputs: 1 },
  },
  {
    id: "error-handler",
    type: "error-handler",
    name: "Error Handler",
    description: "Handle errors in workflow",
    icon: "AlertTriangle",
    defaultData: {
      label: "Error Handler",
      catchAll: true,
      retryCount: 3,
      retryDelay: 5,
    },
    category: "advanced",
    color: "red",
    ports: { inputs: 1, outputs: 2 }, // Success and error outputs
  },
];

// Organize nodes into categories
export const nodeCategories: NodeCategory[] = [
  {
    id: "triggers",
    name: "Triggers",
    description: "Start your workflow",
    icon: "Zap",
    nodes: nodeTemplates.filter((n) => n.category === "triggers"),
  },
  {
    id: "logic",
    name: "Logic",
    description: "Control flow with conditions",
    icon: "Filter",
    nodes: nodeTemplates.filter((n) => n.category === "logic"),
  },
  {
    id: "actions",
    name: "Actions",
    description: "Perform operations",
    icon: "Settings",
    nodes: nodeTemplates.filter((n) => n.category === "actions"),
  },
  {
    id: "control",
    name: "Control Flow",
    description: "Manage execution flow",
    icon: "Layers",
    nodes: nodeTemplates.filter((n) => n.category === "control"),
  },
  {
    id: "advanced",
    name: "Advanced",
    description: "Variables and error handling",
    icon: "Code",
    nodes: nodeTemplates.filter((n) => n.category === "advanced"),
  },
];
