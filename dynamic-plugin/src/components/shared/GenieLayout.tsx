/* eslint-disable @typescript-eslint/ban-ts-comment */
import React, { ReactNode, useState, useEffect } from 'react';
import Helmet from 'react-helmet';
import { AIStateProvider } from '@redhat-cloud-services/ai-react-state';
import { stateManager } from '../utils/aiStateManager';
import '../genie.css';
// Import PatternFly ChatBot CSS as the last import to override styles
import '@patternfly/chatbot/dist/css/main.css';
// Import react-grid-layout CSS
import 'react-grid-layout/css/styles.css';

interface GenieLayoutProps {
  title: string;
  children: ReactNode;
  mainContent?: ReactNode;
}

// Test connection to health endpoint for status display
async function testConnection(): Promise<{ success: boolean; message: string }> {
  try {
    const healthEndpoints = ['/readiness'];

    for (const endpoint of healthEndpoints) {
      try {
        const response = await fetch(`http://localhost:8080${endpoint}`, {
          method: 'GET',
          headers: {
            Accept: 'application/json',
          },
        });

        if (response.ok) {
          return {
            success: true,
            message: `✅ Successfully connected to lightspeed-stack service at localhost:8080${endpoint}`,
          };
        }
      } catch (e) {
        continue;
      }
    }

    return {
      success: false,
      message: `⚠️ Lightspeed-stack service may be running but health endpoints not accessible. You can still try sending queries to /v1/query.`,
    };
  } catch (error) {
    return {
      success: false,
      message: `❌ Cannot connect to lightspeed-stack service at localhost:8080. Please ensure the service is running.`,
    };
  }
}

export function GenieLayout({ title, children, mainContent }: GenieLayoutProps) {
  const [connectionStatus, setConnectionStatus] = useState<{
    success: boolean;
    message: string;
    loading: boolean;
  }>({ success: false, message: 'Testing connection...', loading: true });

  // Test connection when component mounts
  useEffect(() => {
    testConnection()
      .then((result) => {
        setConnectionStatus({
          success: result.success,
          message: result.message,
          loading: false,
        });
      })
      .catch(() => {
        setConnectionStatus({
          success: false,
          message: '❌ Connection test failed',
          loading: false,
        });
      });
  }, []);

  return (
    <AIStateProvider stateManager={stateManager}>
      <div className="genie-standalone">
        {/* @ts-ignore - React 17 compatibility with react-helmet */}
        <Helmet>
          <title>{title}</title>
          <meta name="viewport" content="width=device-width, initial-scale=1" />
        </Helmet>

        {/* Main Content Area */}
        <main className="genie-main">
          {mainContent}
          <div className="genie-container">
            <div className="genie-content">{children}</div>
          </div>
        </main>

        {/* Pinned Status at Bottom */}
        <div className="genie-status-bottom">
          <div className="genie-container">
            <div className="genie-status">
              <p>
                <strong>📡 Health Check:</strong> <code>localhost:8080/readiness</code>
                <span
                  className={`genie-health-status ${
                    connectionStatus.loading
                      ? 'loading'
                      : connectionStatus.success
                      ? 'success'
                      : 'error'
                  }`}
                >
                  {connectionStatus.loading
                    ? '🔄 Testing...'
                    : connectionStatus.success
                    ? '✅ Connected'
                    : '❌ Failed'}
                </span>
              </p>
            </div>
          </div>
        </div>
      </div>
    </AIStateProvider>
  );
}
