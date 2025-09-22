import { useMemo, useEffect } from 'react';
import { useMessages } from '@redhat-cloud-services/ai-react-state';
import { LightSpeedCoreAdditionalProperties } from '@redhat-cloud-services/lightspeed-client';
import {
  CreateDashboardResponse,
  DashboardWidget,
  ManipulateWidgetResponse,
} from '../types/dashboard';
import {
  isCreateDashboardEvent,
  parseCreateDashboardEvent,
  isManipulateWidgetEvent,
  parseManipulateWidgetEvent,
  isManipulateWidgetArgumentsEvent,
  parseManipulateWidgetArgumentsEvent,
  isAddWidgetEvent,
  parseAddWidgetEvent,
} from '../services/eventParser';
import { useManipulationExecutor } from './useManipulationExecutor';
import { useWidgetState } from './useWidgetState';

export function useDashboards() {
  const messages = useMessages<LightSpeedCoreAdditionalProperties>();
  const manipulationExecutor = useManipulationExecutor();

  const { dashboards, baseWidgets, manipulations } = useMemo(() => {
    const allDashboards: CreateDashboardResponse[] = [];
    const allManipulations: ManipulateWidgetResponse[] = [];
    const widgetMap = new Map<string, DashboardWidget>();
    const processedEventIds = new Set<string>();

    messages.forEach((msg) => {
      const toolCalls = msg.additionalAttributes?.toolCalls || [];

      toolCalls.forEach((toolCall) => {
        // Skip events with empty or invalid tokens
        if (!(toolCall as any)?.data?.token || typeof (toolCall as any).data.token !== 'object' || !(toolCall as any).data.token.tool_name) {
          return;
        }

        // Create a unique ID for this specific tool call event
        const hasResponse = !!(toolCall as any).data.token.response;
        const eventId = `${(toolCall as any).data.id}-${(toolCall as any).data.token.tool_name}-${hasResponse ? 'response' : 'arguments'}`;

        // Skip if we've already processed this event
        if (processedEventIds.has(eventId)) {
          return;
        }
        processedEventIds.add(eventId);

        if (isCreateDashboardEvent(toolCall)) {
          const dashboardResponse = parseCreateDashboardEvent(toolCall);
          if (dashboardResponse) {
            allDashboards.push(dashboardResponse);
            // Add widgets to map, using their ID as key (if widgets exist)
            if (dashboardResponse.widgets) {
              dashboardResponse.widgets.forEach((widget) => {
                widgetMap.set(widget.id, widget);
              });
            }
          }
        } else if (isManipulateWidgetEvent(toolCall)) {
          const manipulateResponse = parseManipulateWidgetEvent(toolCall);
          if (manipulateResponse) {
            allManipulations.push(manipulateResponse);
          }
        } else if (isManipulateWidgetArgumentsEvent(toolCall)) {
          const manipulationArgs = parseManipulateWidgetArgumentsEvent(toolCall);
          console.log({ manipulationArgs });
          if (manipulationArgs) {
            // Find the widget and update its position directly
            const existingWidget = widgetMap.get(manipulationArgs.widgetId);
            if (existingWidget) {
              const updatedWidget = {
                ...existingWidget,
                position: {
                  ...existingWidget.position,
                  x: manipulationArgs.position.x,
                  y: manipulationArgs.position.y,
                  ...(manipulationArgs.position.w && { w: manipulationArgs.position.w }),
                  ...(manipulationArgs.position.h && { h: manipulationArgs.position.h }),
                }
              };
              widgetMap.set(manipulationArgs.widgetId, updatedWidget);
            }
          }
        } else if (isAddWidgetEvent(toolCall)) {
          const addWidgetResponse = parseAddWidgetEvent(toolCall);
          console.log({ addWidgetResponse });
          if (addWidgetResponse && addWidgetResponse.widgets) {
            // Add all widgets from the response (usually just one)
            addWidgetResponse.widgets.forEach(widget => {
              // Use the widget exactly as provided by the response, including its ID and position
              widgetMap.set(widget.id, widget);
            });
          }
        }
      });
    });

    return {
      dashboards: allDashboards,
      baseWidgets: Array.from(widgetMap.values()),
      manipulations: allManipulations,
    };
  }, [
    JSON.stringify(
      messages.map((m) => ({
        id: m.id,
        toolCallIds: m.additionalAttributes?.toolCalls?.map((tc) => (tc as any).data?.id) || [],
      })),
    ),
  ]); // Only re-run when message IDs or toolCall IDs change

  // Manage widget state with manipulation support
  const widgetState = useWidgetState(baseWidgets);

  // Update widgets when base widgets change (including empty arrays for new dashboards)
  useEffect(() => {
    widgetState.updateWidgetPositions(baseWidgets);
  }, [baseWidgets.length, widgetState.updateWidgetPositions]); // Only depend on length and the function

  // Add new manipulations to the executor (only add new ones)
  useEffect(() => {
    // Only process the latest manipulation to avoid infinite loops
    if (manipulations.length > 0) {
      const latestManipulation = manipulations[manipulations.length - 1];
      manipulationExecutor.addManipulation(latestManipulation);
    }
  }, [manipulations.length, manipulationExecutor]); // Only depend on length change

  // Get the most recent dashboard as the active one
  const activeDashboard = useMemo(() => {
    return dashboards.length > 0 ? dashboards[dashboards.length - 1] : null;
  }, [dashboards]);

  return {
    dashboards,
    widgets: widgetState.widgets,
    activeDashboard,
    manipulations,
    manipulationExecutor,
    widgetState,
    hasDashboards: dashboards.length > 0,
    hasManipulations: manipulations.length > 0,
  };
}
