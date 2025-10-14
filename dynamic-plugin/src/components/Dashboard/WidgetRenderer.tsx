import React from 'react';
import { DashboardWidget } from '../../types/dashboard';

const componentMapper = {
  TimeSeriesChart: React.lazy(() => import('../PersesBindings/PersesWidgets/PersesTimeSeries')),
  Table: React.lazy(() => import('../PersesBindings/PersesWidgets/PersesTable')),
  PieChart: React.lazy(() => import('../PersesBindings/PersesWidgets/PersesPieChart')),
};

interface WidgetRendererProps {
  widget: DashboardWidget;
}

export function WidgetRenderer({ widget }: WidgetRendererProps) {
  const Component = componentMapper[widget?.componentType];
  if (Component) {
    return (
      <React.Suspense fallback={<div>Loading widget...</div>}>
        <Component {...widget.props} />
      </React.Suspense>
    );
  }

  // Fallback rendering for unknown widget types
  const renderWidgetContent = () => {
    switch (widget.componentType) {
      case 'text':
        return (
          <div>
            {widget.props.title && (
              <h3 style={{ margin: '0 0 10px 0', fontSize: '16px', fontWeight: 'bold' }}>
                {widget.props.title}
              </h3>
            )}
            <p style={{ margin: 0, fontSize: '14px', lineHeight: '1.4' }}>
              {widget.props.content || 'No content available'}
            </p>
          </div>
        );

      case 'ngui':
        return (
          <div>
            {widget.props.title && (
              <h3 style={{ margin: '0 0 10px 0', fontSize: '16px', fontWeight: 'bold' }}>
                {widget.props.title}
              </h3>
            )}
            <p style={{ margin: 0, fontSize: '14px', lineHeight: '1.4' }}>
              {widget.props.content || 'No content available'}
            </p>
          </div>
        );
      default:
        return (
          <div>
            <h4 style={{ margin: '0 0 10px 0', fontSize: '14px', color: '#666' }}>
              Unknown Widget Type: {widget.componentType}
            </h4>
            <pre
              style={{
                fontSize: '12px',
                backgroundColor: '#f8f9fa',
                padding: '10px',
                borderRadius: '4px',
                overflow: 'auto',
                maxHeight: '200px',
              }}
            >
              {JSON.stringify(widget, null, 2)}
            </pre>
          </div>
        );
    }
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {renderWidgetContent()}
    </div>
  );
}
