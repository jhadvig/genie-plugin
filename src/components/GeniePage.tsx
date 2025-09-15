import * as React from 'react';
import { useState, useCallback, useRef, useEffect, useMemo } from 'react';
import Helmet from 'react-helmet';
import { useTranslation } from 'react-i18next';
import './genie.css';

// Types based on ai-web-clients interfaces
interface Message {
  id: string;
  content: string;
  role: 'user' | 'assistant';
  timestamp: Date;
  streaming?: boolean;
}

interface AIClient {
  sendMessage: (content: string, options?: { stream?: boolean }) => Promise<string>;
}

// Real AI Client that follows the ai-web-clients pattern
class LightspeedClient implements AIClient {
  private baseURL: string;
  private fetchFunction: (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;
  
  constructor(config: { baseURL: string; fetchFunction?: typeof fetch }) {
    this.baseURL = config.baseURL.startsWith('http') ? config.baseURL : `http://${config.baseURL}`;
    this.fetchFunction = config.fetchFunction || fetch;
  }

  async sendMessage(content: string, options?: { stream?: boolean }): Promise<string> {
    console.log(`[LightspeedClient] Sending message to ${this.baseURL}:`, content);
    
    try {
      const requestBody = {
        query: content,
        // Remove additional fields that might cause 422 errors
        // conversation_id: this.generateConversationId(),
        // model: 'genie-assistant',
        // stream: options?.stream || false,
      };

      console.log(`[LightspeedClient] Request body:`, requestBody);

      const response = await this.fetchFunction(`${this.baseURL}/v1/query`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify(requestBody),
      });

      console.log(`[LightspeedClient] Response status:`, response.status, response.statusText);

      if (!response.ok) {
        // Get the error details from the response
        let errorDetails = '';
        try {
          const errorData = await response.json();
          errorDetails = JSON.stringify(errorData, null, 2);
          console.error(`[LightspeedClient] Error response:`, errorData);
        } catch (e) {
          errorDetails = await response.text();
          console.error(`[LightspeedClient] Error response text:`, errorDetails);
        }
        throw new Error(`HTTP ${response.status}: ${response.statusText}\nDetails: ${errorDetails}`);
      }

      const data = await response.json();
      console.log(`[LightspeedClient] Response data:`, data);
      
      // Handle different response formats that your API might return
      if (data.response) {
        return data.response;
      } else if (data.message) {
        return data.message;
      } else if (data.answer) {
        return data.answer;
      } else if (typeof data === 'string') {
        return data;
      } else {
        console.warn('Unexpected response format:', data);
        return data.content || 'Received response from AI service';
      }
      
    } catch (error) {
      console.error('[LightspeedClient] Error sending message:', error);
      
      // Provide helpful error messages based on the error type
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new Error(`Cannot connect to AI service at ${this.baseURL}. Please ensure the service is running.`);
      } else if (error instanceof Error) {
        throw new Error(`AI service error: ${error.message}`);
      } else {
        throw new Error('Unknown error occurred while communicating with AI service');
      }
    }
  }

//   private generateConversationId(): string {
//     // Generate a simple conversation ID (in production, this might be managed by the state manager)
//     return `conv_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
//   }

  // Test connection to AI service
  async testConnection(): Promise<{ success: boolean; message: string }> {
    try {
      // Try the health endpoint first, fallback to root if needed
      const healthEndpoints = ['/health', '/v1/health', '/readiness', '/'];
      
      for (const endpoint of healthEndpoints) {
        try {
          const response = await this.fetchFunction(`${this.baseURL}${endpoint}`, {
            method: 'GET',
            headers: {
              'Accept': 'application/json',
            },
          });

          if (response.ok) {
            return { 
              success: true, 
              message: `✅ Successfully connected to lightspeed service at ${this.baseURL}${endpoint}` 
            };
          }
        } catch (e) {
          // Continue to next endpoint
          continue;
        }
      }
      
      return { 
        success: false, 
        message: `⚠️ Lightspeed service is running but health endpoints not accessible. You can still try sending queries.` 
      };
    } catch (error) {
      return { 
        success: false, 
        message: `❌ Cannot connect to lightspeed service at ${this.baseURL}. Please ensure the service is running.` 
      };
    }
  }
}

// State manager following ai-web-clients pattern
class ChatStateManager {
  private messages: Message[] = [];
  private listeners: Set<() => void> = new Set();
  private client: AIClient;
  private isInitialized: boolean = false;

  constructor(client: AIClient) {
    this.client = client;
  }

  // Initialize method following ai-web-clients pattern
  async init(): Promise<void> {
    if (this.isInitialized) return;
    
    try {
      // Test connection first (if client supports it)
      if (typeof (this.client as any).testConnection === 'function') {
        const connectionTest = await (this.client as any).testConnection();
        
        if (connectionTest.success) {
          console.log('[ChatStateManager]', connectionTest.message);
          this.addMessage({
            content: "Hello! I'm Genie, your AI assistant. How can I help you with OpenShift today?",
            role: 'assistant',
          });
        } else {
          console.warn('[ChatStateManager]', connectionTest.message);
          this.addMessage({
            content: `${connectionTest.message}\n\nYou can still send messages, and I'll attempt to connect when you do.`,
            role: 'assistant',
          });
        }
      } else {
        // Fallback if no test method available
        this.addMessage({
          content: "Hello! I'm Genie, your AI assistant. How can I help you with OpenShift today?",
          role: 'assistant',
        });
      }
      
      this.isInitialized = true;
      console.log('[ChatStateManager] Initialized successfully');
    } catch (error) {
      console.error('[ChatStateManager] Initialization failed:', error);
      this.addMessage({
        content: "I'm having trouble connecting to the AI service. Please check that the service is running on localhost:8080.",
        role: 'assistant',
      });
    }
  }

  addMessage(message: Omit<Message, 'id' | 'timestamp'>): Message {
    const newMessage: Message = {
      ...message,
      id: Math.random().toString(36).substr(2, 9),
      timestamp: new Date(),
    };
    
    this.messages.push(newMessage);
    this.notifyListeners();
    return newMessage;
  }

  async sendMessage(content: string, options?: { stream?: boolean }): Promise<void> {
    // Add user message
    this.addMessage({ content, role: 'user' });

    // Add streaming assistant message
    const assistantMessage = this.addMessage({ 
      content: '', 
      role: 'assistant', 
      streaming: true 
    });

    try {
      const response = await this.client.sendMessage(content, options);
      
      // Update the assistant message
      const messageIndex = this.messages.findIndex(m => m.id === assistantMessage.id);
      if (messageIndex >= 0) {
        this.messages[messageIndex] = {
          ...assistantMessage,
          content: response,
          streaming: false,
        };
        this.notifyListeners();
      }
    } catch (error) {
      // Handle error by updating the message with specific error information
      const messageIndex = this.messages.findIndex(m => m.id === assistantMessage.id);
      if (messageIndex >= 0) {
        let errorMessage = 'Sorry, I encountered an error. Please try again.';
        
        if (error instanceof Error) {
          if (error.message.includes('Cannot connect')) {
            errorMessage = `⚠️ Cannot connect to AI service at localhost:8080. Please ensure your AI service is running and accessible.`;
          } else if (error.message.includes('HTTP error')) {
            errorMessage = `⚠️ ${error.message}. Please check your AI service configuration.`;
          } else {
            errorMessage = `⚠️ ${error.message}`;
          }
        }
        
        this.messages[messageIndex] = {
          ...assistantMessage,
          content: errorMessage,
          streaming: false,
        };
        this.notifyListeners();
      }
      
      console.error('[ChatStateManager] Error in sendMessage:', error);
    }
  }

  getMessages(): Message[] {
    return [...this.messages];
  }

  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  private notifyListeners(): void {
    this.listeners.forEach(listener => listener());
  }
}

// React hooks following ai-web-clients pattern
function useMessages(stateManager: ChatStateManager): Message[] {
  const [messages, setMessages] = useState<Message[]>(stateManager.getMessages());

  useEffect(() => {
    const unsubscribe = stateManager.subscribe(() => {
      setMessages(stateManager.getMessages());
    });
    return unsubscribe;
  }, [stateManager]);

  return messages;
}

function useSendMessage(stateManager: ChatStateManager) {
  return useCallback((content: string, options?: { stream?: boolean }) => {
    return stateManager.sendMessage(content, options);
  }, [stateManager]);
}

// Chat Interface Component
function ChatInterface({ stateManager }: { stateManager: ChatStateManager }) {
  const { t } = useTranslation('plugin__genie-plugin');
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  
  const messages = useMessages(stateManager);
  const sendMessage = useSendMessage(stateManager);

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
      await sendMessage(messageContent, { stream: true });
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
        {messages.map((message) => (
          <div 
            key={message.id} 
            className={`genie-message ${message.role === 'assistant' ? 'genie-message--assistant' : 'genie-message--user'}`}
          >
            <div className="genie-message-content">
              {message.content || (message.streaming ? t('Thinking...') : '')}
              {message.streaming && (
                <span className="genie-typing-indicator">
                  <span></span><span></span><span></span>
                </span>
              )}
            </div>
            <div className="genie-message-time">
              {message.timestamp.toLocaleTimeString()}
            </div>
          </div>
        ))}
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

// Main Genie Page Component
export default function GeniePage() {
  const { t } = useTranslation('plugin__genie-plugin');
  const [connectionStatus, setConnectionStatus] = useState<{
    success: boolean;
    message: string;
    loading: boolean;
  }>({ success: false, message: 'Testing connection...', loading: true });

  // Create state manager (following ai-web-clients pattern)
  const stateManager = useMemo(() => {
    // Create actual lightspeed client (following ai-web-clients Option A pattern)
    const client = new LightspeedClient({
      baseURL: 'localhost:8080',
      fetchFunction: (input, init) => fetch(input, init)
    });
    
    // Create state manager (init will be called by provider-like pattern)
    const manager = new ChatStateManager(client);
    
    // Initialize async - once resolved, the client is ready
    manager.init();
    
    // Test connection and update status
    if (typeof (client as any).testConnection === 'function') {
      (client as any).testConnection().then((result: { success: boolean; message: string }) => {
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
    } else {
      setConnectionStatus({
        success: false,
        message: '⚠️ Connection test not available',
        loading: false
      });
    }
    
    return manager;
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
                <small>{t('Ensure your lightspeed service is running and accessible. Check browser console for detailed request/response logs.')}</small>
              </p>
            </div>

            <ChatInterface stateManager={stateManager} />
          </div>
        </div>
      </main>
    </div>
  );
}
