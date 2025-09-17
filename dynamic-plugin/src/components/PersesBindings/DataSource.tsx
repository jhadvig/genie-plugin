import * as React from 'react';
import { DatasourceSelect } from '@perses-dev/plugin-system';
import { DEFAULT_PROM, PROM_DATASOURCE_KIND } from '@perses-dev/prometheus-plugin';

const DataSource = () => {
  return (
    <DatasourceSelect
      datasourcePluginKind={PROM_DATASOURCE_KIND}
      value={DEFAULT_PROM}
      onChange={console.log}
      label="Prometheus Datasource"
    />
  );
};

export default DataSource;
