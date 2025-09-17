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

// Initialize state manager outside React scope (following Red Hat Cloud Services pattern)
const client = new LightspeedClient({
  baseUrl: 'http://localhost:8080',
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
      await sendMessage(message, { stream: true });
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
  return (
    <AIStateProvider stateManager={stateManager}>
      <ChatInterface />
    </AIStateProvider>
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
    <div className="genie-standalone">
      {/* @ts-ignore - React 17 compatibility with react-helmet */}
      <Helmet>
        <title>{t('Genie - AI Assistant')}</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
      </Helmet>

      {/* Centered Title */}
      <div className="genie-title-container">
        <div className="genie-title-content">
          <h1 className="genie-title">{t('Genie')}</h1>
          <p className="genie-subtitle">{t('Your AI Assistant for OpenShift')}</p>
        </div>
      </div>

      {/* Chat Area */}
      <main className="genie-main">
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
  );
}
