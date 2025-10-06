import { DashboardWidget } from "src/types/dashboard";

export class DashboardMCPClient {
  private baseURL: string;
  private requestId = 0;
  private sessionId: string | null = null;

  constructor(baseURL: string = 'http://localhost:9081/mcp') {
    this.baseURL = baseURL;
  }

  async initialize(): Promise<void> {
    if (this.sessionId) return;

    const response = await fetch(this.baseURL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: this.requestId++,
        method: 'initialize',
        params: {
          protocolVersion: '2024-11-05',
          capabilities: {},
          clientInfo: {
            name: 'dashboard-frontend-client',
            version: '1.0.0',
          },
        },
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to initialize MCP client: ${response.statusText}`);
    }

    // Extract session ID from response header
    this.sessionId = response.headers.get('Mcp-Session-Id');
    if (!this.sessionId) {
      throw new Error('Server did not return a session ID');
    }

    const result = await response.json();
    if (result.error) {
      throw new Error(`MCP initialization error: ${result.error.message}`);
    }
  }

  private async callTool<T>(name: string, args: Record<string, any>): Promise<T> {
    if (!this.sessionId) {
      await this.initialize();
    }
    const response = await fetch(this.baseURL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Mcp-Session-Id': this.sessionId!,
      },
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: this.requestId++,
        method: "tools/call",
        params: {
          name: name,
          arguments: args,
        },
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to call tool ${name}: ${response.statusText}`);
    }

    const result = await response.json();

    if (result.error) {
      throw new Error(`MCP Error: ${result.error.message}`);
    }

    const content = result?.result?.content?.[0] || result?.result;
    const text  = content?.text;
    if (text) {
      return JSON.parse(text) as T;
    }
    return content;
  }

  async getActiveDashboard(): Promise<getActiveDashboardResponse> {
    return await this.callTool('get_active_dashboard', {})
  }

  async getDashboard(dashboardId: string): Promise<getDashboardResponse> {
    return await this.callTool('get_dashboard', { dashboard_id: dashboardId })
  }
}

export type getActiveDashboardResponse = {
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

export type getDashboardResponse = {
  success: boolean;
  operation: string;
  dashboardId: string;
  message: string;
  timestamp: string;
  dashboard: DashboardAnalysis;
};