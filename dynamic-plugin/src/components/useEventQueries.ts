import { LightSpeedCoreAdditionalProperties } from '@redhat-cloud-services/lightspeed-client';
import {
  isToolExecutionToken,
  isValidToolArguments,
  // mockToolCalls,
  ToolExecutionToken,
} from './mockedToolCalls';
import { useMemo } from 'react';
import { useMessages } from '@redhat-cloud-services/ai-react-state';

// Example prompts
// Can you show me pod CPU usage for the last 6 hours?

function useEventQueries() {
  const messages = useMessages<LightSpeedCoreAdditionalProperties>();
  // mocking the data for now so we don't have to constantly talk to backend
  // const messages: {
  //   additionalAttributes: LightSpeedCoreAdditionalProperties;
  // }[] = useMemo(
  //   () => [
  //     {
  //       additionalAttributes: {
  //         toolCalls: mockToolCalls.toolCalls,
  //       },
  //     },
  //   ],
  //   [mockToolCalls],
  // );
  const tools = useMemo(
    () =>
      messages.reduce<Map<string, ToolExecutionToken>>((acc, msg) => {
        const toolCalls = msg.additionalAttributes.toolCalls?.filter((tc) => {
          if (isToolExecutionToken((tc as any).data)) {
            return true;
          }

          return false;
        });
        if (toolCalls && toolCalls.length > 0) {
          toolCalls.forEach((tc) => {
            if (typeof (tc as any).data.token.arguments === 'object') {
              const token = (tc as any).data.token as ToolExecutionToken;
              if (isValidToolArguments(token.arguments)) {
                const key = JSON.stringify(token);
                if (!acc.has(key)) {
                  // mocking the time until the MCP gives correct timestamps
                  token.arguments.start = '2025-09-18T05:00:00Z';
                  token.arguments.end = '2025-09-18T06:00:00Z';
                  acc.set(key, token);
                }
              }
            }
          });
        }
        return acc;
      }, new Map<string, ToolExecutionToken>()),
    [messages],
  );
  return Array.from(tools.values());
}

export default useEventQueries;
