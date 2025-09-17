import { DatasourceResource, DatasourceSelector, GlobalDatasourceResource } from '@perses-dev/core';
import { DatasourceApi } from '@perses-dev/dashboards';
import { TFunction } from 'i18next';
import { PERSES_PROXY_BASE_PATH } from './perses-client';

export class OcpDatasourceApi implements DatasourceApi {
  constructor(public t: TFunction, public basePath: string) {}
  /**
   * Helper function for getting a proxy URL from separate input parameters.
   * Give the following output according to the definition or not of the input.
   * - /proxy/globaldatasources/{name}
   * - /proxy/projects/{project}/datasources/{name}
   * - /proxy/projects/{project}/dashboards/{dashboard}/{name}
   *
   * NB: despite the fact it's possible, it is useless to give a dashboard without a project as
   * the url will for sure correspond to nothing.
   * @param name
   * @param dashboard
   * @param project
   */
  buildProxyUrl({
    project,
    dashboard,
    name,
  }: {
    project?: string;
    dashboard?: string;
    name: string;
  }): string {
    let url = `${!project && !dashboard ? 'globaldatasources' : 'datasources'}/${encodeURIComponent(
      name,
    )}`;
    if (dashboard) {
      url = `dashboards/${encodeURIComponent(dashboard)}/${url}`;
    }
    if (project) {
      url = `projects/${encodeURIComponent(project)}/${url}`;
    }
    return `${this.basePath}/proxy/${url}`;
  }

  getDatasource = async (
    project: string,
    selector: DatasourceSelector,
  ): Promise<DatasourceResource | undefined> => {
    return Promise.resolve(undefined);
  };

  getGlobalDatasource = async (
    selector: DatasourceSelector,
  ): Promise<GlobalDatasourceResource | undefined> => {
    // For new return always the default data source
    const globalDatasource: GlobalDatasourceResource = {
      kind: 'GlobalDatasource',
      metadata: {
        name: 'default datasource',
      },
      spec: {
        default: true,
        plugin: {
          kind: 'PrometheusDatasource',
          spec: {
            directUrl: PERSES_PROXY_BASE_PATH,
          },
        },
      },
    };
    return Promise.resolve(globalDatasource);
  };

  listDatasources(project: string, pluginKind?: string): Promise<DatasourceResource[]> {
    // We are not switching between datasources
    return Promise.resolve([]);
  }

  listGlobalDatasources(pluginKind?: string): Promise<GlobalDatasourceResource[]> {
    // We are not switching between datasources
    return Promise.resolve([]);
  }
}
