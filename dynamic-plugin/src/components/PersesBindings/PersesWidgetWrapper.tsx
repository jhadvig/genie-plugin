import React, { PropsWithChildren, useMemo } from 'react';

import { DatasourceStoreProvider, VariableProvider } from '@perses-dev/dashboards';
import { ChartsProvider } from '@perses-dev/components';
import { PluginRegistry, TimeRangeProvider } from '@perses-dev/plugin-system';
import { generateChartsTheme, getTheme } from '@perses-dev/components';
import { QueryClientProvider } from '@tanstack/react-query';
import * as prometheusPlugin from '@perses-dev/prometheus-plugin';

import { pluginLoader } from './persesPluginsLoader';
import persesQueryClient from './perses/persesQueryClient';
import { useTranslation } from 'react-i18next';
import { CachedDatasourceAPI } from './CachedDataSource';
import { OcpDatasourceApi } from './persesDataSourceApi';
import { PERSES_PROXY_BASE_PATH } from './perses-client';

export const muiTheme = getTheme('light');
export const chartsTheme = generateChartsTheme(muiTheme, {});

const persesTimeRange = {
  pastDuration: '1h' as prometheusPlugin.DurationString,
};

const PersesWidgetWrapper = ({ children }: PropsWithChildren<Record<string, unknown>>) => {
  const { t } = useTranslation('plugin__genie-plugin');
  const datasourceApi = useMemo(() => {
    return new CachedDatasourceAPI(new OcpDatasourceApi(t, PERSES_PROXY_BASE_PATH));
  }, [t]);
  return (
    <ChartsProvider chartsTheme={chartsTheme}>
      <PluginRegistry pluginLoader={pluginLoader}>
        <QueryClientProvider client={persesQueryClient}>
          <TimeRangeProvider timeRange={persesTimeRange}>
            <VariableProvider>
              <DatasourceStoreProvider datasourceApi={datasourceApi}>
                <div style={{ width: '500px', height: '200px' }}>{children}</div>
              </DatasourceStoreProvider>
            </VariableProvider>
          </TimeRangeProvider>
        </QueryClientProvider>
      </PluginRegistry>
    </ChartsProvider>
  );
};

export default PersesWidgetWrapper;
