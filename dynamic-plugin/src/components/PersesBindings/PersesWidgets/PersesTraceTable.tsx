import { Panel } from '@perses-dev/dashboards';
import React from 'react';
import PersesWidgetWrapper from '../PersesWidgetWrapper';
import { DataQueriesProvider } from '@perses-dev/plugin-system';

type PersesTraceTableProps = {
  query: string;
};

const PersesTraceTable = ({ query }: PersesTraceTableProps) => {
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
              query,
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
                kind: 'TraceTable',
                spec: {},
              },
            },
          }}
        />
      </DataQueriesProvider>
    </PersesWidgetWrapper>
  );
};

export default PersesTraceTable;
