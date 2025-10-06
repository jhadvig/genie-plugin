import { useMemo, useEffect, useState } from 'react';
import { useStreamChunk } from '@redhat-cloud-services/ai-react-state';
import { LightSpeedCoreAdditionalProperties } from '@redhat-cloud-services/lightspeed-client';
import { CreateDashboardResponse, DashboardWidget } from '../types/dashboard';
import {
  isCreateDashboardEvent,
  parseCreateDashboardEvent,
  isManipulateWidgetArgumentsEvent,
  parseManipulateWidgetArgumentsEvent,
  isAddWidgetEvent,
  parseAddWidgetEvent,
} from '../services/eventParser';
import { DashboardMCPClient } from '../services/dashboardMCPClient';

export function useDashboards() {
  const streamChunk = useStreamChunk<LightSpeedCoreAdditionalProperties>();
  console.log('streamChunk', streamChunk);
  const [dashboards, setDashboards] = useState<CreateDashboardResponse[]>([]);
  const [widgets, setWidgets] = useState<DashboardWidget[]>([]);
  const [dashboardMCPClient] = useState(() => new DashboardMCPClient());

  function handleToolCalls(toolCalls: any[]) {
    toolCalls.forEach((toolCall) => {
      // Skip events with empty or invalid tokens
      if (
        !(toolCall as any)?.data?.token ||
        typeof (toolCall as any).data.token !== 'object' ||
        !(toolCall as any).data.token.tool_name
      ) {
        return;
      }

      if (isCreateDashboardEvent(toolCall)) {
        const dashboardResponse = parseCreateDashboardEvent(toolCall);
        if (dashboardResponse) {
          setDashboards((prev) => [...prev, dashboardResponse]);
          setWidgets(dashboardResponse.widgets ?? []);
        }
      } else if (isManipulateWidgetArgumentsEvent(toolCall)) {
        const manipulationArgs = parseManipulateWidgetArgumentsEvent(toolCall);
        if (manipulationArgs) {
          // Find the widget and update its position directly
          setWidgets((prev) =>
            prev.map((w) =>
              w.id === manipulationArgs.widgetId
                ? {
                    ...w,
                    position: {
                      ...w.position,
                      x: manipulationArgs.position.x,
                      y: manipulationArgs.position.y,
                      ...(manipulationArgs.position.w && { w: manipulationArgs.position.w }),
                      ...(manipulationArgs.position.h && { h: manipulationArgs.position.h }),
                    },
                  }
                : w,
            ),
          );
        }
      } else if (isAddWidgetEvent(toolCall)) {
        const addWidgetResponse = parseAddWidgetEvent(toolCall);
        if (addWidgetResponse && addWidgetResponse.widgets) {
          // Add all widgets from the response (usually just one)
          setWidgets((prev) => [...prev, ...(addWidgetResponse.widgets ?? [])]);
        }
      }
    });
  }

  useEffect(() => {
    if (streamChunk && streamChunk.additionalAttributes?.toolCalls) {
      console.log('streamChunk.additionalAttributes.toolCalls', streamChunk.additionalAttributes.toolCalls);
      handleToolCalls(streamChunk.additionalAttributes.toolCalls);
    }
  }, [streamChunk]);

  // Fetch active dashboard from MCP when no dashboards from messages
  const activeDashboard = useMemo(() => {
      // const fetchDashboard = async () => {
      //   try {
      //     const activeDashboardResponse = await dashboardMCPClient.getActiveDashboard();
      //     console.log('Active dashboard response:', activeDashboardResponse);
      //     // widgetState.updateWidgetPositions(activeDashboardResponse.analysis.widgets);
      //   } catch (error) {
      //     console.error('Error fetching active dashboard:', error);
      //   }
      // };
      return dashboards.length > 0 ? dashboards[dashboards.length - 1] : null;
  }, [dashboards.length, dashboardMCPClient]);

  return {
    dashboards,
    widgets,
    activeDashboard,
    hasDashboards: dashboards.length > 0,
  };
}
