import React, { useMemo } from 'react';
import {
  K8sResourceCommon,
  ResourceLink,
  RowProps,
  TableColumn,
  TableData,
  Timestamp,
  VirtualizedTable,
  useK8sWatchResource,
} from '@openshift-console/dynamic-plugin-sdk';
import { ColumnSpec, DataTypeConfig } from './dataTypeRegistry';

type NGUIField = {
  id: string;
  name: string;
  data_path?: string;
};

type ResourceListWidgetProps = {
  config?: DataTypeConfig;
  fields: NGUIField[];
  k8s?: {
    group?: string;
    version?: string;
    kind?: string;
    namespaced?: boolean;
    namespace?: string;
  };
};

const getByPath = (obj: any, path?: string): any => {
  if (!obj || !path) return undefined;
  const parts = path.split('.');
  let current = obj;
  for (const p of parts) {
    if (current == null) return undefined;
    current = current[p];
  }
  return current;
};

const matchFieldToColumnSpec = (field: NGUIField, allSpecs: ColumnSpec[]): ColumnSpec | undefined => {
  const byId = allSpecs.find((s) => s.id.toLowerCase() === field.id.toLowerCase());
  if (byId) return byId;
  const byName = allSpecs.find((s) => s.title.toLowerCase() === field.name.toLowerCase());
  return byName;
};

const ResourceListWidget: React.FC<ResourceListWidgetProps> = ({ config, fields, k8s }) => {
  const groupVersionKind = {
    group: k8s?.group,
    version: k8s?.version,
    kind: k8s?.kind,
  };

  // TODO: Confirm if this is the correct TS type for the items
  const [items, loaded, loadError] = useK8sWatchResource<K8sResourceCommon[]>({
    groupVersionKind,
    isList: true,
    namespaced: k8s?.namespaced ,
    namespace: k8s?.namespace,
  });

  console.log(
    'items', items,
    'loaded', loaded,
    'loadError', loadError,
  );

  const hasLiveData = loaded && !loadError;

  const tableData: K8sResourceCommon[] = hasLiveData ? (items || []) : [];

  // Find the column specs for the fields
  const resolvedColumnSpecs: ColumnSpec[] = useMemo(() => {
    const allSpecs = config?.columnSpecs ?? [];
    return fields
      .map((field) => matchFieldToColumnSpec(field, allSpecs))
      .filter((spec): spec is ColumnSpec => Boolean(spec));
  }, [fields, config?.columnSpecs]);

  // Create the table columns from the column specs
  const tableColumns: TableColumn<K8sResourceCommon>[] = useMemo(
    () =>
      resolvedColumnSpecs.map((columnSpec) => ({
        title: columnSpec.title,
        id: columnSpec.id,
      })),
    [resolvedColumnSpecs],
  );

  const Row: React.FC<RowProps<K8sResourceCommon>> = ({ obj, activeColumnIDs }) => {
    return (
      <>
        {resolvedColumnSpecs.map((columnSpec) => {
          let content: React.ReactNode = null;
          if (columnSpec.accessor) {
            content = columnSpec.accessor(obj);
          } else if (columnSpec.type === 'resource') {
            content = (
              <ResourceLink
                groupVersionKind={groupVersionKind}
                name={getByPath(obj, columnSpec.path)}
                namespace={obj?.metadata?.namespace}
              />
            );
          } else if (columnSpec.type === 'date') {
            const ts = getByPath(obj, columnSpec.path);
            content = ts ? <Timestamp timestamp={ts} /> : '-';
          } else {
            const v = getByPath(obj, columnSpec.path);
            content = v ?? '-';
          }
          return (
            <TableData key={columnSpec.id} id={columnSpec.id} activeColumnIDs={activeColumnIDs}>
              {content}
            </TableData>
          );
        })}
      </>
    );
  };

  return (
    <div style={{ marginTop: '12px' }}>
      <VirtualizedTable<K8sResourceCommon>
        data={tableData}
        unfilteredData={tableData}
        loaded={loaded}
        loadError={loadError}
        columns={tableColumns}
        Row={Row}
        aria-label="Resource list"
      />
    </div>
  );
};

export default ResourceListWidget;


