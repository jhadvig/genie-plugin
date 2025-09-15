import * as React from 'react';
import { useState, useRef, useEffect } from 'react';
import Helmet from 'react-helmet';
import { useTranslation } from 'react-i18next';
import { AIStateProvider, useSendMessage, useMessages } from '@redhat-cloud-services/ai-react-state';
import { createClientStateManager } from '@redhat-cloud-services/ai-client-state';
import { LightspeedClient } from '@redhat-cloud-services/lightspeed-client';
import './genie.css';

// Initialize state manager outside React scope (following Red Hat Cloud Services pattern)
const client = new LightspeedClient({ 
  baseUrl: 'http://localhost:8080', 
  fetchFunction: (input, init) => fetch(input, init),
});

const stateManager = createClientStateManager(client);

// Initialize immediately when module loads (no longer auto-creates conversations)
stateManager.init().then(() => {
  console.log('[Genie] State manager initialized successfully');
}).catch((error) => {
  console.error('[Genie] State manager initialization failed:', error);
});

// ChatInterface using official Red Hat Cloud Services hooks
function ChatInterface() {
  const { t } = useTranslation('plugin__genie-plugin');
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [welcomeShown, setWelcomeShown] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  
  const sendMessage = useSendMessage();
  const messages = useMessages();

  // Show welcome message if no messages exist
  useEffect(() => {
    if (!welcomeShown && messages.length === 0) {
      // Add welcome message to show initial state
      setWelcomeShown(true);
    }
  }, [messages, welcomeShown]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async () => {
    if (!inputValue.trim() || isLoading) return;

    const messageContent = inputValue.trim();
    setInputValue('');
    setIsLoading(true);

    try {
      await sendMessage(messageContent, { stream: false });
    } finally {
      setIsLoading(false);
    }
  };


  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="genie-chat">
      <div className="genie-messages">
        {messages.length === 0 ? (
          <div className="genie-message genie-message--assistant">
            <div className="genie-message-content">
              {t("Hello! I'm Genie, your AI assistant. How can I help you with OpenShift today?")}
            </div>
            <div className="genie-message-time">
              {new Date().toLocaleTimeString()}
            </div>
          </div>
        ) : (
          messages.map((msg) => {
            const message = msg as any; // Type assertion to handle Red Hat Cloud Services message format
            const isAssistant = message.role === 'bot' || message.role === 'assistant';
            return (
              <div 
                key={msg.id} 
                className={`genie-message ${isAssistant ? 'genie-message--assistant' : 'genie-message--user'}`}
              >
                <div className="genie-message-content">
                  {message.answer || message.query || message.message || message.content || ''}
                </div>
                <div className="genie-message-time">
                  {new Date(message.timestamp || message.createdAt || Date.now()).toLocaleTimeString()}
                </div>
              </div>
            );
          })
        )}
        <div ref={messagesEndRef} />
      </div>
      
      <div className="genie-input-area">
        <div className="genie-input-wrapper">
          <input
            type="text"
            className="genie-input"
            placeholder={t('Ask me anything about OpenShift...')}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyPress={handleKeyPress}
            disabled={isLoading}
          />
          <button 
            className="genie-send-button"
            onClick={handleSend}
            disabled={!inputValue.trim() || isLoading}
          >
            {isLoading ? t('Sending...') : t('Send')}
          </button>
        </div>
      </div>
    </div>
  );
}

// Test connection to health endpoint for status display
async function testConnection(): Promise<{ success: boolean; message: string }> {
  try {
    const healthEndpoints = ['/health', '/v1/health', '/readiness', '/'];
    
    for (const endpoint of healthEndpoints) {
      try {
        const response = await fetch(`http://localhost:8080${endpoint}`, {
          method: 'GET',
          headers: {
            'Accept': 'application/json',
          },
        });

        if (response.ok) {
          return { 
            success: true, 
            message: `✅ Successfully connected to lightspeed-stack service at localhost:8080${endpoint}` 
          };
        }
      } catch (e) {
        continue;
      }
    }
    
    return { 
      success: false, 
      message: `⚠️ Lightspeed-stack service may be running but health endpoints not accessible. You can still try sending queries to /v1/query.` 
    };
  } catch (error) {
    return { 
      success: false, 
      message: `❌ Cannot connect to lightspeed-stack service at localhost:8080. Please ensure the service is running.` 
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
    testConnection().then(result => {
      setConnectionStatus({
        success: result.success,
        message: result.message,
        loading: false
      });
    }).catch(() => {
      setConnectionStatus({
        success: false,
        message: '❌ Connection test failed',
        loading: false
      });
    });
  }, []);

  return (
    <div className="genie-standalone">
      <Helmet>
        <title>{t('Genie - AI Assistant')}</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
      </Helmet>
      
      <header className="genie-header">
        <div className="genie-container">
          <h1 className="genie-title">{t('Genie')}</h1>
          <p className="genie-subtitle">{t('Your AI Assistant for OpenShift')}</p>
        </div>
      </header>

      <main className="genie-main">
        <div className="genie-container">
          <div className="genie-content">
            <div className="genie-welcome">
              <h2>{t('Welcome to Genie')}</h2>
              <p>
                {t('This is a standalone AI-powered chat interface for OpenShift assistance. Built using the ai-web-clients architecture pattern.')}
              </p>
              <p>
                <strong>🔗 API Endpoint:</strong> <code>localhost:8080/v1/query</code><br/>
                <strong>📡 Health Check:</strong> <code>localhost:8080/readiness</code> 
                <span 
                  className={`genie-health-status ${connectionStatus.loading ? 'loading' : connectionStatus.success ? 'success' : 'error'}`}
                >
                  {connectionStatus.loading ? '🔄 Testing...' : connectionStatus.success ? '✅ Connected' : '❌ Failed'}
                </span>
                <br/>
                <small>{t('Ensure your lightspeed-stack service is running and accessible. Check browser console for detailed request/response logs.')}</small>
              </p>
            </div>

            <App />
          </div>
        </div>
      </main>
    </div>
  );
}
