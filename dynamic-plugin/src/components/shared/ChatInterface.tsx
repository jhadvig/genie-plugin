import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  useSendMessage,
  useMessages,
} from '@redhat-cloud-services/ai-react-state';
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

interface ChatInterfaceProps {
  welcomeTitle?: string;
  welcomeDescription?: string;
  placeholder?: string;
}

export function ChatInterface({
  welcomeTitle,
  welcomeDescription,
  placeholder
}: ChatInterfaceProps) {
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
    <Chatbot displayMode={ChatbotDisplayMode.embedded}>
      <ChatbotContent>
        <ChatbotWelcomePrompt
          title={welcomeTitle || t("Hello! I'm Genie")}
          description={welcomeDescription || t('Your AI assistant for OpenShift. Ask me anything!')}
        />
        <MessageBox>{formatMessages()}</MessageBox>
      </ChatbotContent>
      <ChatbotFooter>
        <MessageBar
          onSendMessage={handlePatternFlySend}
          placeholder={placeholder || t('Ask me anything about OpenShift...')}
          hasMicrophoneButton={false}
          isSendButtonDisabled={isLoading}
          alwayShowSendButton={true}
        />
      </ChatbotFooter>
    </Chatbot>
  );
}