import { dynamicImportPluginLoader } from '@perses-dev/plugin-system';
import * as prometheusPlugin from '@perses-dev/prometheus-plugin';
import * as timeseriesChartPlugin from '@perses-dev/timeseries-chart-plugin';
import * as timeSeriesTablePlugin from '@perses-dev/timeseries-table-plugin';
import * as pieChartPlugin from '@perses-dev/pie-chart-plugin';
import * as scatterChartPlugin from '@perses-dev/scatter-chart-plugin';
import * as statusHistoryChartPlugin from '@perses-dev/status-history-chart-plugin';
import * as markdownPlugin from '@perses-dev/markdown-plugin';
import * as tablePlugin from '@perses-dev/table-plugin';
import * as staticListVariablePlugin from '@perses-dev/static-list-variable-plugin';
import * as tempoPlugin from '@perses-dev/tempo-plugin';
import * as traceTablePlugin from '@perses-dev/trace-table-plugin';
import * as tracingGanttChartPlugin from '@perses-dev/tracing-gantt-chart-plugin';
/**
 * A PluginLoader that includes all the "built-in" plugins that are bundled with Perses by default and additional custom plugins
 */
export const pluginLoader = dynamicImportPluginLoader([
  {
    resource: timeseriesChartPlugin.getPluginModule(),
    importPlugin: async () => Promise.resolve(timeseriesChartPlugin),
  },
  {
    resource: prometheusPlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(prometheusPlugin),
  },
  {
    resource: pieChartPlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(pieChartPlugin),
  },
  {
    resource: scatterChartPlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(scatterChartPlugin),
  },
  {
    resource: timeSeriesTablePlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(timeSeriesTablePlugin),
  },
  {
    resource: statusHistoryChartPlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(statusHistoryChartPlugin),
  },
  {
    resource: markdownPlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(markdownPlugin),
  },
  {
    resource: staticListVariablePlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(staticListVariablePlugin),
  },
  {
    resource: tablePlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(tablePlugin),
  },
  {
    resource: tempoPlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(tempoPlugin),
  },
  {
    resource: traceTablePlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(traceTablePlugin),
  },
  {
    resource: tracingGanttChartPlugin.getPluginModule(),
    importPlugin: () => Promise.resolve(tracingGanttChartPlugin),
  },
]);
