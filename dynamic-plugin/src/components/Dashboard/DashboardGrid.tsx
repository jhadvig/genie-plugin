/* eslint-disable @typescript-eslint/ban-ts-comment */
import React from 'react';
import ReactGridLayout from 'react-grid-layout';
import { DashboardWidget } from '../../types/dashboard';
import { WidgetRenderer } from './WidgetRenderer';

interface DashboardGridProps {
  widgets: DashboardWidget[];
  cols?: number;
  rowHeight?: number;
  width?: number;
  isDraggable?: boolean;
  isResizable?: boolean;
  onLayoutChange?: (layout: any[]) => void;
}

export function DashboardGrid({
  widgets,
  cols = 12,
  rowHeight = 60,
  width = 1200,
  isDraggable = true,
  isResizable = true,
  onLayoutChange,
}: DashboardGridProps) {
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
