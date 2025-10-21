/**
 * Environment detection utilities for genie-plugin
 */

export interface EnvironmentConfig {
  baseUrl: string;
  isDevelopment: boolean;
  isConsoleMode: boolean;
}

/**
 * Detects the current environment and returns appropriate configuration
 */
export function getEnvironmentConfig(): EnvironmentConfig {
  // Check if we're running in the OpenShift Console
  const isConsoleMode = window.location.origin.includes('localhost:9000') || 
                       window.location.pathname.includes('/console/') ||
                       window.location.hostname !== 'localhost';

  // Check if we're in development mode (webpack dev server)
  const isDevelopment = window.location.origin.includes('localhost:9001') && !isConsoleMode;

  let baseUrl: string;

  if (isConsoleMode) {
    // Use console proxy when running through OpenShift Console
    baseUrl = '/api/proxy/plugin/genie-plugin/lightspeed';
  } else if (isDevelopment) {
    // Use webpack proxy when running in development mode
    baseUrl = '';
  } else {
    // Fallback to console proxy for production
    baseUrl = '/api/proxy/plugin/genie-plugin/lightspeed';
  }

  return {
    baseUrl,
    isDevelopment,
    isConsoleMode,
  };
}

/**
 * Logs the current environment configuration for debugging
 */
export function logEnvironmentConfig(): void {
  const config = getEnvironmentConfig();
  
  console.log('[Genie] Environment Configuration:', {
    baseUrl: config.baseUrl,
    isDevelopment: config.isDevelopment,
    isConsoleMode: config.isConsoleMode,
    currentOrigin: window.location.origin,
    currentPathname: window.location.pathname,
  });
}

/**
 * Gets the appropriate base URL for the LightspeedClient
 */
export function getLightspeedBaseUrl(): string {
  return getEnvironmentConfig().baseUrl;
}

