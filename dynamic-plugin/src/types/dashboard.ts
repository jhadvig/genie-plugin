export type DashboardWidget = {
  id: string;
  componentType: 'text' | 'chart' | 'ngui' | 'log';
  position: {
    x: number;
    y: number;
    w: number;
    h: number;
  };
  props: Record<string, any> & {
    // For chart widgets, we might have a specific Perses component
    persesComponent?: string;
    ngui_id?: string;
    ngui_content?: string;
  };
  // Optional breakpoint for responsive layouts
  breakpoint: string;
  description?: string;
};

export type DashboardLayout = {
  id: string;
  layoutId: string;
  name: string;
  description: string;
};

export type CreateDashboardResponse = {
  success: boolean;
  operation: string;
  activeLayoutId: string;
  message: string;
  timestamp: string;
  widgets?: DashboardWidget[]; // Optional since empty dashboards don't have widgets
  totalFound?: number; // Optional for backward compatibility
  layout: DashboardLayout;
};

export type CreateDashboardEvent = {
  event: 'tool_call';
  data: {
    id: number;
    role: 'tool_execution';
    token: {
      tool_name: 'create_dashboard';
      response: CreateDashboardResponse;
    };
  };
};

export type WidgetChange = {
  widgetId: string;
  action: 'moved' | 'repositioned' | 'resized' | 'removed' | 'added';
  breakpoint: string;
  wasTargeted: boolean;
  reason: string;
  previousState: {
    h: number;
    w: number;
    x: number;
    y: number;
  };
  newState: {
    h: number;
    w: number;
    x: number;
    y: number;
  };
};

export type ManipulateWidgetResponse = {
  success: boolean;
  operation: string;
  layoutId: string;
  targetedWidgets: string[];
  allChanges: WidgetChange[];
  summary: {
    totalAffected: number;
    targeted: number;
    collateralChanges: number;
    operations: Record<string, number>;
    reasons: Record<string, number>;
  };
  message: string;
  affectedBreakpoints: string[];
  timestamp: string;
  widgets: DashboardWidget[];
};

export type ManipulateWidgetEvent = {
  event: 'tool_call';
  data: {
    id: number;
    role: 'tool_execution';
    token: {
      tool_name: 'manipulate_widget';
      response: ManipulateWidgetResponse;
    };
  };
};

export type FindWidgetsResponse = {
  success: boolean;
  operation: string;
  activeLayoutId: string;
  message: string;
  timestamp: string;
  widgets: DashboardWidget[];
};

export type AddWidgetResponse = {
  success: boolean;
  operation: string;
  activeLayoutId: string;
  message: string;
  timestamp: string;
  widgets: DashboardWidget[];
};

export type AddWidgetEvent = {
  event: 'tool_call';
  data: {
    id: number;
    role: 'tool_execution';
    token: {
      tool_name: 'add_widget';
      response: AddWidgetResponse;
    };
  };
};

export type GenerateUIEvent = {
  event: 'tool_result';
  data: {
    id: number;
    token: {
      tool_name: 'generate_ui';
      response: string;
      artifact: string;
      status: string;
    };
  };
};

export type ManipulateWidgetArgumentsEvent = {
  event: 'tool_call';
  data: {
    id: number;
    role: 'tool_execution';
    token: {
      tool_name: 'manipulate_widget';
      arguments: {
        x: string;
        y: string;
        widget_id: string;
        operation: string;
        w?: string; // Optional width
        h?: string; // Optional height
      };
    };
  };
};

export type ActiveDashboardResponse = {
  success: boolean;
  operation: string;
  activeLayoutId: string;
  message: string;
  timestamp: string;
  analysis: DashboardAnalysis;
};

export type DashboardAnalysis = {
  breakpoint: string;
  cols: number;
  description: string;
  globalConstraints: GlobalConstraints;
  layoutId: string;
  name: string;
  widgets: DashboardWidget[];
};

export type GlobalConstraints = {
  maxItems: number;
  defaultItemSize: {
    w: number;
    h: number;
  };
};

export type Dashboard = {
  id: string;
  name: string;
  description: string;
  widgets: DashboardWidget[];
};

export type ListDashboardsResponse = {
  success: boolean;
  operation: string;
  message: string;
  timestamp: string;
  layouts: Array<{
    id: string;
    layoutId: string;
    name: string;
    description: string;
  }>;
};
