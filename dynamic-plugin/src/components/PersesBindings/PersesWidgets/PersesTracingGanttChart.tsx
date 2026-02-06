import { Panel } from '@perses-dev/dashboards';
import React from 'react';
import PersesWidgetWrapper from '../PersesWidgetWrapper';
import { DataQueriesProvider } from '@perses-dev/plugin-system';

type PersesTracingGanttChartProps = {
  traceId: string;
};

const PersesTracingGanttChart = ({ traceId }: PersesTracingGanttChartProps) => {
  return (
    <PersesWidgetWrapper>
      <DataQueriesProvider
        definitions={[
          {
            kind: 'TempoTraceQuery',
            spec: {
              datasource: {
                kind: 'TempoDatasource',
              },
              query: traceId,
            },
          },
        ]}
      >
        <Panel
          panelOptions={{
            hideHeader: true,
          }}
          definition={{
            kind: 'Panel',
            spec: {
              display: { name: '' },
              plugin: {
                kind: 'TracingGanttChart',
                spec: {},
              },
            },
          }}
        />
      </DataQueriesProvider>
    </PersesWidgetWrapper>
  );
};

export default PersesTracingGanttChart;
