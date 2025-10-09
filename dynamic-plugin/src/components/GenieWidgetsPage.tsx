import React, { useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation } from 'react-router-dom';
import { useDashboards } from '../hooks/useDashboards';
import { DashboardGrid } from './Dashboard';
import { ChatInterface, GenieLayout } from './shared';
import { DashboardMCPClient } from '../services/dashboardClient';
import './utils/reactPolyfills';

// Dashboard Layout component
function DashboardLayout() {
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const dashboardId = searchParams.get('dashboardId');

  const { widgets, activeDashboard, hasDashboards } = useDashboards(dashboardId || undefined);
  const dashboardClient = useRef(new DashboardMCPClient());
  const saveTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleLayoutChange = useCallback(
    (layout: any[]) => {
      if (!activeDashboard?.layout?.layoutId) return;

      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }

      saveTimeoutRef.current = setTimeout(async () => {
        try {
          await dashboardClient.current.updateWidgetPositions(layout);
          console.log('Layout saved successfully');
        } catch (error) {
          console.error('Failed to save layout:', error);
        }
      }, 1000); // wait 1 second after user stops dragging to save the layout
    },
    [activeDashboard?.layout?.layoutId],
  );

  return (
    <div style={{ padding: '20px' }}>
      {activeDashboard && activeDashboard.layout && (
        <div style={{ marginBottom: '20px' }}>
          <h2 style={{ margin: '0 0 10px 0', fontSize: '24px', fontWeight: 'bold' }}>
            {activeDashboard.layout.name || 'Untitled Dashboard'}
          </h2>
          <p style={{ margin: '0 0 20px 0', color: '#666', fontSize: '14px' }}>
            {activeDashboard.layout.description || 'No description available'}
          </p>
        </div>
      )}

      <DashboardGrid
        widgets={widgets}
        onLayoutChange={handleLayoutChange}
        cols={12}
        rowHeight={60}
        width={1200}
        isDraggable={true}
        isResizable={true}
      />

      {hasDashboards && (
        <div
          style={{
            marginTop: '20px',
            padding: '15px',
            backgroundColor: '#f8f9fa',
            borderRadius: '4px',
            fontSize: '14px',
            color: '#666',
          }}
        >
          <strong>Dashboard Info:</strong> {widgets.length} widget(s) loaded
          {activeDashboard && (
            <>
              <br />
              <strong>Dashboard ID:</strong> {activeDashboard.layout.layoutId}
              <br />
              <strong>Active Layout ID:</strong> {activeDashboard.activeLayoutId}
            </>
          )}
        </div>
      )}
    </div>
  );
}

// Main Genie Widgets Page Component
export default function GenieWidgetsPage() {
  const { t } = useTranslation('plugin__genie-plugin');

  return (
    <GenieLayout title={t('Genie Widgets - AI Dashboard Assistant')}>
      {/* This is a temporary layout to display the chat interface and the dashboard side by side */}
      <div className="chat-interface">
        <ChatInterface
          welcomeTitle={t("Hello! I'm Genie!")}
          welcomeDescription={t('An AI assistant for OpenShift.')}
          placeholder={t('Message Genie...')}
        />
      </div>
      <div className="dashboard">
        <DashboardLayout />
      </div>
    </GenieLayout>
  );
}

// Can you create a new dashboard witch a chart of a cluster CPU usage over the last 15 minutes? Check what we have available for queries and use the best that firts for a time series chart.

// Can you add a new widget that shows memory suage per namesapce?
