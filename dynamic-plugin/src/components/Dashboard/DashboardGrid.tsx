/* eslint-disable @typescript-eslint/ban-ts-comment */
import React, { useEffect, useRef } from 'react';
import ReactGridLayout from 'react-grid-layout';
import { DashboardWidget } from '../../types/dashboard';
import { WidgetRenderer } from './WidgetRenderer';
import { useManipulationExecutor } from '../../hooks/useManipulationExecutor';
import { useWidgetState } from '../../hooks/useWidgetState';

interface DashboardGridProps {
  widgets: DashboardWidget[];
  manipulationExecutor?: ReturnType<typeof useManipulationExecutor>;
  widgetState?: ReturnType<typeof useWidgetState>;
  cols?: number;
  rowHeight?: number;
  width?: number;
  isDraggable?: boolean;
  isResizable?: boolean;
  onLayoutChange?: (layout: any[]) => void;
}

export function DashboardGrid({
  widgets,
  manipulationExecutor,
  widgetState,
  cols = 12,
  rowHeight = 60,
  width = 1200,
  isDraggable = true,
  isResizable = true,
  onLayoutChange,
}: DashboardGridProps) {
  const executedManipulationsRef = useRef<Set<string>>(new Set());

  // Execute pending manipulations
  useEffect(() => {
    if (!manipulationExecutor || !widgetState) return;

    const nextManipulation = manipulationExecutor.getNextManipulation();
    if (!nextManipulation) return;

    // Prevent re-execution of the same manipulation
    if (executedManipulationsRef.current.has(nextManipulation.id)) {
      return;
    }

    console.log('Executing manipulation:', nextManipulation);

    // Mark as executed immediately in our ref
    executedManipulationsRef.current.add(nextManipulation.id);

    // Apply the manipulation by updating widget positions
    const { manipulation } = nextManipulation;

    if (manipulation.widgets) {
      // Update widget state with new positions
      widgetState.updateWidgetPositions(manipulation.widgets);

      // Mark manipulation as executed in the executor
      manipulationExecutor.executeManipulation(nextManipulation.id);

      // Trigger layout change callback if provided
      if (onLayoutChange) {
        const newLayout = manipulation.widgets.map((widget) => ({
          i: widget.id,
          x: widget.position.x,
          y: widget.position.y,
          w: widget.position.w,
          h: widget.position.h,
        }));
        onLayoutChange(newLayout);
      }
    }
  }, [
    manipulationExecutor?.hasPendingManipulations,
    widgetState,
    onLayoutChange,
    manipulationExecutor,
  ]);

  // Create layout configuration from widgets
  const layout = widgets.map((widget) => ({
    i: widget.id,
    x: widget.position.x,
    y: widget.position.y,
    w: widget.position.w,
    h: widget.position.h,
    minW: 2,
    minH: 2,
  }));

  // Create grid items
  const gridItems = widgets.map((widget) => (
    <div
      key={widget.id}
      style={{
        border: '1px solid #ccc',
        borderRadius: '4px',
        padding: '10px',
        backgroundColor: '#f9f9f9',
        overflow: 'hidden',
      }}
    >
      <WidgetRenderer widget={widget} />
    </div>
  ));

  if (widgets.length === 0) {
    return (
      <div
        style={{
          padding: '20px',
          textAlign: 'center',
          color: '#666',
          border: '2px dashed #ccc',
          borderRadius: '8px',
          margin: '20px 0',
        }}
      >
        <h3>No Widgets Yet</h3>
        <p>This dashboard is empty. Ask Genie to add widgets to your dashboard!</p>
      </div>
    );
  }

  return (
    // @ts-ignore
    <ReactGridLayout
      className="layout"
      layout={layout}
      cols={cols}
      rowHeight={rowHeight}
      width={width}
      isDraggable={isDraggable}
      isResizable={isResizable}
      margin={[16, 16]}
      onLayoutChange={onLayoutChange}
    >
      {gridItems}
    </ReactGridLayout>
  );
}
