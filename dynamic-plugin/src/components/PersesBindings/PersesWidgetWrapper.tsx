import React, { PropsWithChildren, useMemo } from 'react';

import { DatasourceStoreProvider, VariableProvider } from '@perses-dev/dashboards';
import { ChartsProvider, SnackbarProvider, typography } from '@perses-dev/components';
import { PluginRegistry, RouterProvider, TimeRangeProvider } from '@perses-dev/plugin-system';
import { generateChartsTheme, getTheme } from '@perses-dev/components';
import { QueryClientProvider } from '@tanstack/react-query';
import * as prometheusPlugin from '@perses-dev/prometheus-plugin';

import { pluginLoader } from './persesPluginsLoader';
import persesQueryClient from './perses/persesQueryClient';
import { useTranslation } from 'react-i18next';
import { CachedDatasourceAPI } from './CachedDataSource';
import { OcpDatasourceApi } from './persesDataSourceApi';
import { PERSES_PROXY_BASE_PATH } from './perses-client';
import { ThemeProvider } from '@mui/material';
import { Link as RouterLink, useNavigate } from 'react-router-dom-v5-compat';

export const muiTheme = getTheme('light', {
  typography: {
    ...typography,
    fontFamily: 'var(--pf-t--global--font--family--body)',
  },
  cssVariables: true,
  // Remove casting once https://github.com/perses/perses/pull/3443 is merged
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
} as any);
export const chartsTheme = generateChartsTheme(muiTheme, {});

const persesTimeRange = {
  pastDuration: '1h' as prometheusPlugin.DurationString,
};

const PersesWidgetWrapper = ({ children }: PropsWithChildren<Record<string, unknown>>) => {
  const { t } = useTranslation('plugin__genie-plugin');
  const navigate = useNavigate();
  const datasourceApi = useMemo(() => {
    return new CachedDatasourceAPI(new OcpDatasourceApi(t, PERSES_PROXY_BASE_PATH));
  }, [t]);

  return (
    <ThemeProvider theme={muiTheme}>
      <RouterProvider RouterComponent={RouterLink} navigate={navigate}>
        <ChartsProvider chartsTheme={chartsTheme}>
          <SnackbarProvider
            anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
            variant="default"
          >
            <PluginRegistry pluginLoader={pluginLoader}>
              <QueryClientProvider client={persesQueryClient}>
                <TimeRangeProvider timeRange={persesTimeRange}>
                  <VariableProvider>
                    <DatasourceStoreProvider datasourceApi={datasourceApi}>
                      <div style={{ width: '100%', height: '100%' }}>{children}</div>
                    </DatasourceStoreProvider>
                  </VariableProvider>
                </TimeRangeProvider>
              </QueryClientProvider>
            </PluginRegistry>
          </SnackbarProvider>
        </ChartsProvider>
      </RouterProvider>
    </ThemeProvider>
  );
};

export default PersesWidgetWrapper;
