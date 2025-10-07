import { useEffect, useState } from 'react';
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
import DashboardUtils, { NormalizedDashboard } from '../components/utils/dashboard.utils';

export function useDashboards() {
  const streamChunk = useStreamChunk<LightSpeedCoreAdditionalProperties>();
  console.log('streamChunk', streamChunk);
  const [dashboards, setDashboards] = useState<CreateDashboardResponse[]>([]);
  const [widgets, setWidgets] = useState<DashboardWidget[]>([]);
  const [dashboardMCPClient] = useState(() => new DashboardMCPClient());
  const [activeDashboard, setActiveDashboard] = useState<NormalizedDashboard | undefined>(undefined);

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

  useEffect(() => {
    if (dashboards.length > 0) {
      const lastCreated = dashboards[dashboards.length - 1];
      const normalized = DashboardUtils.normalizeResponse(lastCreated);
      setActiveDashboard(normalized);
    }
  }, [dashboards.length]);

  useEffect(() => {
    async function fetchActive() {
      try {
        if (dashboards.length === 0) {
          const resp = await dashboardMCPClient.getActiveDashboard();
          if (resp) {
            const normalizedActive = DashboardUtils.normalizeResponse(resp);
            setActiveDashboard(normalizedActive);
          }
        }
      } catch (error) {
        console.error('Error fetching active dashboard:', error);
      }
    }
    fetchActive();
  }, [dashboards.length, dashboardMCPClient]);

  return {
    dashboards,
    widgets,
    activeDashboard,
    hasDashboards: dashboards.length > 0,
  };
}
