import * as React from 'react';
import Helmet from 'react-helmet';
import { useTranslation } from 'react-i18next';
import './genie.css';

export default function GeniePage() {
  const { t } = useTranslation('plugin__genie-plugin');

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
                {t('This is a standalone page that runs independently of the OpenShift Console layout. It provides a full-screen experience for the Genie AI assistant.')}
              </p>
            </div>

            <div className="genie-chat">
              <div className="genie-messages">
                <div className="genie-message genie-message--assistant">
                  <div className="genie-message-content">
                    {t('Hello! I\'m Genie, your AI assistant. How can I help you with OpenShift today?')}
                  </div>
                </div>
              </div>
              
              <div className="genie-input-area">
                <div className="genie-input-wrapper">
                  <input
                    type="text"
                    className="genie-input"
                    placeholder={t('Ask me anything about OpenShift...')}
                  />
                  <button className="genie-send-button">
                    {t('Send')}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
