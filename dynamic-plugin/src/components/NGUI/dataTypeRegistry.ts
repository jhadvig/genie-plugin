import { K8sResourceKind } from '@openshift-console/dynamic-plugin-sdk';

export type ColumnSpec = {
  id: string;
  title: string;
  type?: string;
  path?: string;
  // TODO: Assuming K8sResourceKind is the correct type for the object, but I should confirm
  accessor?: (obj: K8sResourceKind) => string | number;
};

export type DataTypeConfig = {
  columnSpecs: ColumnSpec[];
};

export const dataTypeRegistry: Record<string, DataTypeConfig> = {
  pods_list_in_namespace: {
    columnSpecs: [
      // Showcasing how we specify the type as 'resource' to render a link to the resource
      { id: 'NAME', title: 'Name', type: 'resource', path: 'metadata.name' },
      { id: 'STATUS', title: 'Status', path: 'status.phase' },
      {
        id: 'READY',
        title: 'Ready',
        // For derived values, we have to specify the type because we can't get it from the schema
        type: 'string',
        // Showcasing how we can specify an accessor function to calculate a cell's value when it's not a simple, direct lookup
        accessor: (pod: K8sResourceKind) => {
          const statuses = (pod.status?.containerStatuses as any[]) ?? [];
          const total = statuses.length;
          const ready = statuses.filter((s) => s.ready).length;
          return `${ready}/${total}`;
        },
      },
      { id: 'AGE', title: 'Age', type: 'date-time', path: 'metadata.creationTimestamp' },
      { id: 'IP', title: 'IP', type: 'ip', path: 'status.podIP' },
      { id: 'NODE', title: 'Node', type: 'string', path: 'spec.nodeName' },
    ],
  },
};


