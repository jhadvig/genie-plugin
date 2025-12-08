import React, { useMemo } from 'react';
import {
  K8sResourceCommon,
  RowProps,
  TableColumn,
  TableData,
  VirtualizedTable,
  useK8sWatchResource,
  useK8sModel,
  WatchK8sResource,
} from '@openshift-console/dynamic-plugin-sdk';
import { ColumnSpec, DataTypeConfig } from './dataTypeRegistry';
import {
  getByPath,
  getSchemaFormat,
  matchFieldToColumnSpec,
  NGUIField,
  typeRenderers,
} from './widget-utils';

type ResourceListWidgetProps = {
  config?: DataTypeConfig;
  fields: NGUIField[];
  k8s?: WatchK8sResource;
};

const ResourceListWidget: React.FC<ResourceListWidgetProps> = ({ config, fields, k8s }) => {
  const [items, loaded, loadError] = useK8sWatchResource<K8sResourceCommon[]>({
    ...k8s,
    isList: true,
  });
  const { groupVersionKind } = k8s;
  const [model] = useK8sModel(groupVersionKind);
  const hasLiveData = loaded && !loadError;
  const tableData: K8sResourceCommon[] = hasLiveData ? items || [] : [];

  const resolvedColumnSpecs: ColumnSpec[] = useMemo(() => {
    const allSpecs = config?.columnSpecs ?? [];
    return fields
      .map((field) => matchFieldToColumnSpec(field, allSpecs))
      .filter((spec): spec is ColumnSpec => Boolean(spec));
  }, [fields, config?.columnSpecs]);

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
        {resolvedColumnSpecs.map((spec) => {
          const value = spec.accessor ? spec.accessor(obj) : getByPath(obj, spec.path);

          const effectiveType = spec.type ?? getSchemaFormat(model, spec.path);
          
          const typeRenderer = effectiveType ? typeRenderers[effectiveType] : undefined;

          const content = typeRenderer
            ? typeRenderer(value, obj, groupVersionKind)
            : value ?? '-';

          return (
            <TableData key={spec.id} id={spec.id} activeColumnIDs={activeColumnIDs}>
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


