# React Grid Layout Documentation

## Overview

React Grid Layout is "a grid layout system much like Packery or Gridster, for React" that provides draggable and resizable grid layout with responsive breakpoints. It's React-only and does not require jQuery.

## Installation

```bash
npm install react-grid-layout
```

Required stylesheets:
```
/node_modules/react-grid-layout/css/styles.css
/node_modules/react-resizable/css/styles.css
```

## Core Data Structures

### Layout Item Format

```javascript
{
  i: string,        // Item identifier (required, must match React key)
  x: number,        // Grid column position (0-based)
  y: number,        // Grid row position (0-based)
  w: number,        // Width in grid units
  h: number,        // Height in grid units
  static?: boolean, // Cannot be dragged/resized (default: false)
  minW?: number,    // Minimum width in grid units
  maxW?: number,    // Maximum width in grid units
  minH?: number,    // Minimum height in grid units
  maxH?: number,    // Maximum height in grid units
  isDraggable?: boolean,  // Override global draggable setting
  isResizable?: boolean,  // Override global resizable setting
  moved?: boolean,        // Internal: item was moved programmatically
  isBounded?: boolean     // Item is constrained to container bounds
}
```

### Responsive Layouts Structure

```javascript
const layouts = {
  lg: [{ i: "a", x: 0, y: 0, w: 6, h: 2 }],
  md: [{ i: "a", x: 0, y: 0, w: 4, h: 2 }],
  sm: [{ i: "a", x: 0, y: 0, w: 2, h: 2 }],
  xs: [{ i: "a", x: 0, y: 0, w: 1, h: 2 }],
  xxs: [{ i: "a", x: 0, y: 0, w: 1, h: 2 }]
}
```

## Basic GridLayout Component

```javascript
import GridLayout from "react-grid-layout";

<GridLayout
  className="layout"
  layout={layout}
  cols={12}
  rowHeight={30}
  width={1200}
  onLayoutChange={onLayoutChange}
>
  <div key="a">Item A</div>
  <div key="b">Item B</div>
</GridLayout>
```

### GridLayout Props

```typescript
interface GridLayoutProps {
  // Basic props
  className?: string;
  style?: React.CSSProperties;
  width: number;                    // Container width in px

  // Layout configuration
  layout: Layout[];                 // Array of layout items
  cols?: number;                    // Number of columns (default: 12)
  rowHeight?: number;               // Row height in px (default: 150)
  maxRows?: number;                 // Maximum number of rows

  // Interaction
  isDraggable?: boolean;            // Items draggable (default: true)
  isResizable?: boolean;            // Items resizable (default: true)
  isBounded?: boolean;              // Constrain to container (default: false)

  // Layout behavior
  preventCollision?: boolean;       // Prevent item collision (default: false)
  compactType?: 'vertical' | 'horizontal' | null; // Compaction type
  verticalCompact?: boolean;        // Vertical compaction (deprecated)

  // Margins and spacing
  margin?: [number, number];        // [horizontal, vertical] margin (default: [10, 10])
  containerPadding?: [number, number]; // Container padding

  // Resize handles
  resizeHandles?: Array<'s' | 'w' | 'e' | 'n' | 'sw' | 'nw' | 'se' | 'ne'>;
  resizeHandle?: React.ComponentType<any>;

  // Event callbacks
  onLayoutChange?: (layout: Layout[]) => void;
  onDrag?: (layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement) => void;
  onDragStart?: (layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement) => void;
  onDragStop?: (layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement) => void;
  onResize?: (layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement) => void;
  onResizeStart?: (layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement) => void;
  onResizeStop?: (layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement) => void;

  // Child items
  children: React.ReactNode[];
}
```

## Responsive GridLayout

```javascript
import { Responsive as ResponsiveGridLayout } from "react-grid-layout";

<ResponsiveGridLayout
  className="layout"
  layouts={layouts}
  breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
  cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
  rowHeight={60}
  onLayoutChange={onLayoutChange}
  onBreakpointChange={onBreakpointChange}
>
  <div key="a">Item A</div>
  <div key="b">Item B</div>
</ResponsiveGridLayout>
```

### ResponsiveGridLayout Props

Extends GridLayout props with:

```typescript
interface ResponsiveGridLayoutProps extends Omit<GridLayoutProps, 'layout' | 'cols'> {
  // Responsive configuration
  layouts: { [breakpoint: string]: Layout[] };
  breakpoints?: { [breakpoint: string]: number };  // Breakpoint widths
  cols?: { [breakpoint: string]: number };         // Columns per breakpoint

  // Responsive callbacks
  onBreakpointChange?: (newBreakpoint: string, newCols: number) => void;
  onWidthChange?: (containerWidth: number, margin: [number, number], cols: number, containerPadding: [number, number]) => void;
}
```

### Default Breakpoints
```javascript
{
  lg: 1200,
  md: 996,
  sm: 768,
  xs: 480,
  xxs: 0
}
```

### Default Columns
```javascript
{
  lg: 12,
  md: 10,
  sm: 6,
  xs: 4,
  xxs: 2
}
```

## WidthProvider HOC

Automatically provides width to the grid layout:

```javascript
import { Responsive, WidthProvider } from "react-grid-layout";

const ResponsiveGridLayout = WidthProvider(Responsive);

// Usage - no need to provide width prop
<ResponsiveGridLayout
  className="layout"
  layouts={layouts}
  breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
  cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
>
  <div key="a">a</div>
  <div key="b">b</div>
</ResponsiveGridLayout>
```

## Data-Grid Attributes (Alternative)

Instead of layout prop, you can use data-grid attributes:

```javascript
<GridLayout className="layout" cols={12} rowHeight={30} width={1200}>
  <div key="a" data-grid={{ x: 0, y: 0, w: 1, h: 2, static: true }}>a</div>
  <div key="b" data-grid={{ x: 1, y: 0, w: 3, h: 2, minW: 2, maxW: 4 }}>b</div>
</GridLayout>
```

## Event Handler Parameters

### Layout Change Events
```javascript
onLayoutChange(layout: Layout[]): void
```

### Drag Events
```javascript
onDrag(layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement): void
onDragStart(layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement): void
onDragStop(layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement): void
```

### Resize Events
```javascript
onResize(layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement): void
onResizeStart(layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement): void
onResizeStop(layout: Layout[], oldItem: LayoutItem, newItem: LayoutItem, placeholder: LayoutItem, event: MouseEvent, element: HTMLElement): void
```

### Responsive Events
```javascript
onBreakpointChange(newBreakpoint: string, newCols: number): void
onWidthChange(containerWidth: number, margin: [number, number], cols: number, containerPadding: [number, number]): void
```

## Static Imports

For server-side rendering compatibility:

```javascript
import GridLayout from "react-grid-layout/build/GridLayout";
import ResponsiveGridLayout from "react-grid-layout/build/ResponsiveGridLayout";
```

## CSS Classes

The component generates these CSS classes:
- `.react-grid-layout` - Main container
- `.react-grid-item` - Individual grid items
- `.react-grid-item.react-grid-placeholder` - Placeholder during drag
- `.react-resizable-handle` - Resize handles

## Best Practices

1. **Always provide unique keys** that match layout item `i` property
2. **Use WidthProvider** for responsive behavior
3. **Handle layout changes** to persist state
4. **Set appropriate margins** for visual spacing
5. **Use static items** for fixed elements
6. **Implement bounds checking** with `isBounded`
7. **Consider performance** with large layouts

## Performance Notes

- Uses CSS Transform for positioning (hardware accelerated)
- Efficient collision detection
- Optimized for large numbers of items
- Server-side rendering support