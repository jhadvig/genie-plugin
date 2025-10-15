import * as React from 'react';
import { TimeRangeValue, AbsoluteTimeRange, RelativeTimeRange } from '@perses-dev/core';
import { DatasourceStoreProvider, Panel, VariableProvider } from '@perses-dev/dashboards';
import useResizeObserver from 'use-resize-observer';
import { ChartsProvider } from '@perses-dev/components';
import {
  DataQueriesProvider,
  PluginRegistry,
  TimeRangeProvider,
  useSuggestedStepMs,
} from '@perses-dev/plugin-system';
import { useMemo } from 'react';

import { DEFAULT_PROM } from '@perses-dev/prometheus-plugin';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import * as prometheusPlugin from '@perses-dev/prometheus-plugin';
import { pluginLoader } from './persesPluginsLoader';
import { OcpDatasourceApi } from './persesDataSourceApi';
import { useTranslation } from 'react-i18next';
import { getCSRFToken } from '@openshift-console/dynamic-plugin-sdk/lib/utils/fetch/console-fetch-utils';
import { CachedDatasourceAPI } from './CachedDataSource';
import { PERSES_PROXY_BASE_PATH } from './perses-client';

import { generateChartsTheme, getTheme } from '@perses-dev/components';

export const muiTheme = getTheme('light');
export const chartsTheme = generateChartsTheme(muiTheme, {});

// const testEvent = {
//   event: 'tool_call',
//   data: {
//     id: 150,
//     role: 'tool_execution',
//     token: {
//       tool_name: 'execute_range_query',
//       arguments: {
//         duration: '5m',
//         end: '2025-09-18T06:00:00Z',
//         query: 'sum(rate(container_cpu_usage_seconds_total[5m]))',
//         start: '2025-09-18T05:00:00Z',
//         step: '1m',
//       },
//     },
//   },
// };

type TimeSeriesProps = {
  duration: string;
  end: string;
  query: string;
  start: string;
  step: string;
};

console.log('CRSF token:', getCSRFToken());

const useTimeRange = (start?: string, end?: string, duration?: string) => {
  const result = useMemo(() => {
    let timeRange: TimeRangeValue;
    if (start && end) {
      timeRange = {
        start: new Date(start),
        end: new Date(end),
      } as AbsoluteTimeRange;
    } else {
      timeRange = { pastDuration: duration || '1h' } as RelativeTimeRange;
    }
    return timeRange;
  }, [duration, end, start]);
  return result;
};

const TimeSeries = ({ query }: TimeSeriesProps) => {
  const datasource = DEFAULT_PROM;
  const { width, ref: boxRef } = useResizeObserver();
  const suggestedStepMs = useSuggestedStepMs(width);

  const definitions =
    query !== ''
      ? [
          {
            kind: 'PrometheusTimeSeriesQuery',
            spec: {
              datasource: {
                kind: datasource.kind,
                name: datasource.name,
              },
              query: query,
            },
          },
        ]
      : [];

  return (
    <div ref={boxRef} style={{ width: '100%', height: '400px' }}>
      <DataQueriesProvider definitions={definitions} options={{ suggestedStepMs, mode: 'range' }}>
        <Panel
          panelOptions={{
            hideHeader: true,
          }}
          definition={{
            kind: 'Panel',
            spec: {
              queries: [],
              display: { name: '' },
              plugin: {
                kind: 'TimeSeriesChart',
                spec: {
                  visual: {
                    stack: 'all',
                  },
                },
              },
            },
          }}
        />
      </DataQueriesProvider>
    </div>
  );
};

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 0,
      queryFn: async ({ queryKey }) => {
        throw new Error(`No queryFn defined for queryKey: ${queryKey.join(',')}`);
      },
    },
  },
});

const persesTimeRange = {
  pastDuration: '1h' as prometheusPlugin.DurationString,
};

export const MockedTimeSeries = (props: TimeSeriesProps) => {
  const { t } = useTranslation('plugin__genie-plugin');
  const timeSeriesProps = props;
  const timeRange = useTimeRange(
    timeSeriesProps.start,
    timeSeriesProps.end,
    timeSeriesProps.duration,
  );
  const datasourceApi = useMemo(() => {
    return new CachedDatasourceAPI(new OcpDatasourceApi(t, PERSES_PROXY_BASE_PATH));
  }, [t]);
  return (
    <ChartsProvider chartsTheme={chartsTheme}>
      <PluginRegistry pluginLoader={pluginLoader}>
        <QueryClientProvider client={queryClient}>
          <TimeRangeProvider timeRange={persesTimeRange}>
            <VariableProvider>
              <DatasourceStoreProvider datasourceApi={datasourceApi}>
                <div style={{ width: '100%', height: '100%' }}>
                  <TimeRangeProvider timeRange={timeRange} refreshInterval="0s">
                    <TimeSeries {...timeSeriesProps} />
                  </TimeRangeProvider>
                </div>
              </DatasourceStoreProvider>
            </VariableProvider>
          </TimeRangeProvider>
        </QueryClientProvider>
      </PluginRegistry>
    </ChartsProvider>
  );
};
