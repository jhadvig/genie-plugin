import { useState, useCallback, useRef, useEffect } from 'react';
import { DashboardWidget } from '../types/dashboard';

export function useWidgetState(initialWidgets: DashboardWidget[]) {
  const [widgets, setWidgets] = useState<DashboardWidget[]>(initialWidgets);
  const initialWidgetsRef = useRef<DashboardWidget[]>(initialWidgets);

  // Update the ref when initialWidgets change
  useEffect(() => {
    initialWidgetsRef.current = initialWidgets;
  }, [initialWidgets]);

  // Update widget positions while preserving existing props
  const updateWidgetPositions = useCallback((newWidgets: DashboardWidget[]) => {
    setWidgets((prevWidgets) => {
      // Create a map of existing widgets for fast lookup
      const existingWidgetsMap = new Map(prevWidgets.map(widget => [widget.id, widget]));

      // Merge new widgets with existing ones, preserving props
      const mergedWidgets = newWidgets.map(newWidget => {
        const existingWidget = existingWidgetsMap.get(newWidget.id);
        if (existingWidget) {
          // Merge existing widget with new widget, preserving all props
          return {
            ...existingWidget,
            ...newWidget,
            props: {
              ...existingWidget.props,
              ...newWidget.props,
            },
            position: {
              ...existingWidget.position,
              ...newWidget.position,
            }
          };
        }
        // If it's a completely new widget, use it as-is
        return newWidget;
      });

      // Add any existing widgets that weren't in the new list
      const newWidgetIds = new Set(newWidgets.map(w => w.id));
      const remainingWidgets = prevWidgets.filter(widget => !newWidgetIds.has(widget.id));

      return [...mergedWidgets, ...remainingWidgets];
    });
  }, []);

  // Update a single widget position
  const updateWidgetPosition = useCallback(
    (
      widgetId: string,
      newPosition: {
        x: number;
        y: number;
        w: number;
        h: number;
      },
    ) => {
      setWidgets((prev) =>
        prev.map((widget) =>
          widget.id === widgetId
            ? { ...widget, position: { ...widget.position, ...newPosition } }
            : widget,
        ),
      );
    },
    [],
  );

  // Reset widgets to initial state
  const resetWidgets = useCallback(() => {
    setWidgets(initialWidgetsRef.current);
  }, []); // No dependencies since we use ref

  return {
    widgets,
    updateWidgetPositions,
    updateWidgetPosition,
    resetWidgets,
  };
}
