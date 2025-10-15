import React from 'react';
import { DashboardWidget } from '../../types/dashboard';
import DynamicComponent from '@rhngui/patternfly-react-renderer';

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
        if (widget.props.ngui_content) {
          const c = JSON.parse(widget.props.ngui_content);
          console.log(c);
          return <DynamicComponent config={c} />;
        } else {
          return <div>No data</div>;
        }
      case 'log':
        return (
          <div>
            {widget.props.title && (
              <h3 style={{ margin: '0 0 10px 0', fontSize: '16px', fontWeight: 'bold' }}>
                {widget.props.title}
              </h3>
            )}
            <pre>{widget.props.content || 'No content available'}</pre>
          </div>
        );

      case 'chart':
        return (
          <div>
            <h3 style={{ margin: '0 0 10px 0', fontSize: '16px', fontWeight: 'bold' }}>
              {widget.props.title || widget.props.description || 'Chart Widget'}
            </h3>
            <div
              style={{
                backgroundColor: '#e9ecef',
                padding: '20px',
                borderRadius: '4px',
                textAlign: 'left',
                fontSize: '12px',
                color: '#6c757d',
              }}
            >
              <div style={{ textAlign: 'center', marginBottom: '10px' }}>
                📊 {widget.props.persesComponent || 'Chart'} Chart
              </div>
              {widget.props.query && (
                <div style={{ marginBottom: '8px' }}>
                  <strong>Query:</strong>{' '}
                  <code style={{ fontSize: '11px' }}>{widget.props.query}</code>
                </div>
              )}
              {widget.props.duration && (
                <div style={{ marginBottom: '8px' }}>
                  <strong>Duration:</strong> {widget.props.duration}
                </div>
              )}
              {widget.props.step && (
                <div style={{ marginBottom: '8px' }}>
                  <strong>Step:</strong> {widget.props.step}
                </div>
              )}
              <div style={{ textAlign: 'center', marginTop: '10px' }}>
                <small>(PersesTimeSeries component not loaded)</small>
              </div>
            </div>
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
