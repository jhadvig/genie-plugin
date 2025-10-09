import { useEffect, useRef, useState } from 'react';
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
import { DashboardMCPClient } from '../services/dashboardClient';
import DashboardUtils, { NormalizedDashboard } from '../components/utils/dashboard.utils';

export function useDashboards(dashboardId?: string) {
  const streamChunk = useStreamChunk<LightSpeedCoreAdditionalProperties>();
  console.log('streamChunk', streamChunk);
  const [dashboards, setDashboards] = useState<CreateDashboardResponse[]>([]);
  const [widgets, setWidgets] = useState<DashboardWidget[]>([]);
  const dashboardClient = useRef(new DashboardMCPClient());
  const [activeDashboard, setActiveDashboard] = useState<NormalizedDashboard | undefined>(
    undefined,
  );

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

      const toolName = (toolCall as any).data.token.tool_name;
      console.log('Tool called:', toolName);
      console.log('Tool call data:', toolCall);

      if (isCreateDashboardEvent(toolCall)) {
        const dashboardResponse = parseCreateDashboardEvent(toolCall);
        if (dashboardResponse) {
          setDashboards((prev) => [...prev, dashboardResponse]);
          setWidgets(dashboardResponse.widgets ?? []);
        }
      } else if (isManipulateWidgetArgumentsEvent(toolCall)) {
        const manipulationArgs = parseManipulateWidgetArgumentsEvent(toolCall);
        if (manipulationArgs) {
          // Find the widget and update its position directly, ensuring defaults
          setWidgets((prev) => {
            const next = prev.map((w) => {
              if (w.id !== manipulationArgs.widgetId) return w;
              const currentPos = w.position ?? { x: 0, y: 0, w: 4, h: 4 };
              return {
                ...w,
                position: {
                  x: manipulationArgs.position.x ?? currentPos.x,
                  y: manipulationArgs.position.y ?? currentPos.y,
                  w: manipulationArgs.position.w ?? currentPos.w,
                  h: manipulationArgs.position.h ?? currentPos.h,
                },
              } as DashboardWidget;
            });
            return next;
          });
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
      console.log(
        'streamChunk.additionalAttributes.toolCalls',
        streamChunk.additionalAttributes.toolCalls,
      );
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
          const resp = dashboardId
            ? await dashboardClient.current.getDashboard(dashboardId)
            : await dashboardClient.current.getActiveDashboard();
          if (resp) {
            console.log('resp', resp);
            const normalizedActive = DashboardUtils.normalizeResponse(resp);
            console.log('normalizedActive', normalizedActive);
            setActiveDashboard(normalizedActive);
            setWidgets(normalizedActive?.widgets ?? []);
          }
        }
      } catch (error) {
        console.error('Error fetching active dashboard:', error);
      }
    }
    fetchActive();
  }, [dashboards.length, dashboardClient, dashboardId]);

  return {
    dashboards,
    widgets,
    activeDashboard,
    hasDashboards: dashboards.length > 0 || activeDashboard,
  };
}
