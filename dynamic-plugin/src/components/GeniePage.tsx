/* eslint-disable @typescript-eslint/ban-ts-comment */
import * as React from 'react';
import { useState, useEffect, useLayoutEffect, useMemo } from 'react';
import Helmet from 'react-helmet';
import { useTranslation } from 'react-i18next';

async function patchUseId() {
  // @ts-ignore
  const scope = __webpack_share_scopes__?.default;
  if (!scope) {
    return;
  }
  if (scope) {
    let react = await scope.react['*'].get();
    if (!react) {
      return;
    }
    react = react();
    if (!react.useId) {
      console.log('[Genie] Patching React.useId for compatibility');
      react.useId = () => {
        const id = useMemo(() => {
          return crypto.randomUUID();
        }, []);
        return id;
      };
    }
  }
}

patchUseId();

// Polyfill for React 17 compatibility with libraries expecting React 18
if (!(React as any).useInsertionEffect) {
  (React as any).useInsertionEffect = useLayoutEffect;
}
import {
  AIStateProvider,
  useSendMessage,
  useMessages,
} from '@redhat-cloud-services/ai-react-state';
import { createClientStateManager } from '@redhat-cloud-services/ai-client-state';
import { LightspeedClient } from '@redhat-cloud-services/lightspeed-client';
import {
  Chatbot,
  ChatbotDisplayMode,
  ChatbotContent,
  ChatbotWelcomePrompt,
  ChatbotFooter,
  MessageBox,
  Message,
  MessageBar,
} from '@patternfly/chatbot';
import './genie.css';
// Import PatternFly ChatBot CSS as the last import to override styles
import '@patternfly/chatbot/dist/css/main.css';
// Import react-grid-layout CSS
import 'react-grid-layout/css/styles.css';
// import { DatasourceSelect } from '@perses-dev/plugin-system';
// import DataSource from './PersesBindings/DataSource';
import { MockedTimeSeries } from './PersesBindings';
import useEventQueries from './useEventQueries';
import ReactGridLayout from 'react-grid-layout';

// Initialize state manager outside React scope (following Red Hat Cloud Services pattern)
const client = new LightspeedClient({
  baseUrl: 'http://localhost:9001/ols',
  fetchFunction: (input, init) => fetch(input, init),
});

const stateManager = createClientStateManager(client);

// Initialize immediately when module loads (no longer auto-creates conversations)
stateManager
  .init()
  .then(() => {
    console.log('[Genie] State manager initialized successfully');
  })
  .catch((error) => {
    console.error('[Genie] State manager initialization failed:', error);
  });

// ChatInterface using PatternFly ChatBot with Red Hat Cloud Services hooks
function ChatInterface() {
  const { t } = useTranslation('plugin__genie-plugin');
  const [isLoading, setIsLoading] = useState(false);

  const sendMessage = useSendMessage();
  const messages = useMessages();

  // Convert Red Hat Cloud Services messages to PatternFly format
  const formatMessages = () => {
    return messages.map((msg) => {
      const message = msg as any; // Type assertion for Red Hat Cloud Services message format
      const isBot = message.role === 'bot' || message.role === 'assistant';

      return (
        <Message
          key={msg.id}
          name={isBot ? 'Genie' : 'You'}
          role={isBot ? 'bot' : 'user'}
          avatar={
            isBot
              ? 'https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png/openshift.png'
              : 'https://w7.pngwing.com/pngs/831/88/png-transparent-user-profile-computer-icons-user-interface-mystique-miscellaneous-user-interface-design-smile-thumbnail.png'
          }
          timestamp={new Date(
            message.timestamp || message.createdAt || Date.now(),
          ).toLocaleTimeString()}
          content={message.answer || message.query || message.message || message.content || ''}
        />
      );
    });
  };

  const handlePatternFlySend = async (message: string) => {
    if (!message.trim() || isLoading) return;

    setIsLoading(true);
    try {
      await sendMessage(message, { stream: true, requestOptions: {} });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Chatbot displayMode={ChatbotDisplayMode.docked}>
      <ChatbotContent>
        <ChatbotWelcomePrompt
          title={t("Hello! I'm Genie")}
          description={t('Your AI assistant for OpenShift. Ask me anything!')}
        />
        <MessageBox>{formatMessages()}</MessageBox>
      </ChatbotContent>
      <ChatbotFooter>
        <MessageBar
          onSendMessage={handlePatternFlySend}
          placeholder={t('Ask me anything about OpenShift...')}
          hasMicrophoneButton={false}
          isSendButtonDisabled={isLoading}
          alwayShowSendButton={true}
        />
      </ChatbotFooter>
    </Chatbot>
  );
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

// App component following Red Hat Cloud Services pattern
function App() {
  return <ChatInterface />;
}

function Layout() {
  const queryEvents = useEventQueries();
  console.log('Query Events:', queryEvents);

  // Create layout configuration for grid items
  const layout = queryEvents.map((_, index) => ({
    i: `item-${index}`,
    x: (index % 2) * 6, // 2 columns layout
    y: Math.floor(index / 2) * 4, // Row height of 4
    w: 6, // Width of 6 units (half the grid)
    h: 6, // Height of 6 units
    minW: 4,
    minH: 3,
  }));

  // Create grid items
  const gridItems = queryEvents.map((event, index) => {
    return (
      <div
        key={`item-${index}`}
        style={{ border: '1px solid #ccc', padding: '10px', backgroundColor: '#f9f9f9' }}
      >
        <h3>Query: {event.arguments.query}</h3>
        <MockedTimeSeries
          query={event.arguments.query}
          start={event.arguments.start}
          end={event.arguments.end}
          duration={event.arguments.duration}
          step={event.arguments.step}
        />
      </div>
    );
  });

  return (
    // @ts-ignore
    <ReactGridLayout
      className="layout"
      layout={layout}
      cols={12}
      rowHeight={60}
      width={1200}
      isDraggable={true}
      isResizable={true}
      margin={[16, 16]}
    >
      {gridItems}
    </ReactGridLayout>
  );
}

// Main Genie Page Component
export default function GeniePage() {
  const { t } = useTranslation('plugin__genie-plugin');
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
          <title>{t('Genie - AI Assistant')}</title>
          <meta name="viewport" content="width=device-width, initial-scale=1" />
        </Helmet>

        {/* Chat Area */}
        <main className="genie-main">
          <Layout />
          <div className="genie-container">
            <div className="genie-content">
              <App />
            </div>
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
