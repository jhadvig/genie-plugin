import { Panel } from '@perses-dev/dashboards';
import {
  DataQueriesProvider,
  TimeRangeProvider,
  useSuggestedStepMs,
} from '@perses-dev/plugin-system';
import { DEFAULT_PROM } from '@perses-dev/prometheus-plugin';
import React from 'react';
import useResizeObserver from 'use-resize-observer';
import PersesWidgetWrapper from '../PersesWidgetWrapper';
import { useTimeRange } from '../useTimeRange';

type PersesTimeSeriesProps = {
  duration: string;
  end: string;
  query: string;
  start: string;
  step: string;
};

const TimeSeries = ({ query }: PersesTimeSeriesProps) => {
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
    <div ref={boxRef} style={{ width: '100%', height: '100%' }}>
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

const PersesTimeSeries = (props: PersesTimeSeriesProps) => {
  const timeSeriesProps = props;
  const timeRange = useTimeRange(
    timeSeriesProps.start,
    timeSeriesProps.end,
    timeSeriesProps.duration,
  );
  return (
    <PersesWidgetWrapper>
      <TimeRangeProvider timeRange={timeRange} refreshInterval="0s">
        <TimeSeries {...timeSeriesProps} />
      </TimeRangeProvider>
    </PersesWidgetWrapper>
  );
};

export default PersesTimeSeries;
