export type ColumnType = 'string' | 'number' | 'date' | 'status' | 'ip' | 'resource';

export type ColumnSpec = {
  id: string;
  title: string;
  type?: ColumnType;
  path?: string;
  accessor?: (obj: any) => string | number | JSX.Element | null;
};

export type DataTypeConfig = {
  columnSpecs: ColumnSpec[];
};

export const dataTypeRegistry: Record<string, DataTypeConfig> = {
  pods_list_in_namespace: {
    columnSpecs: [
      { id: 'NAME', title: 'Name', type: 'resource', path: 'metadata.name' },
      { id: 'STATUS', title: 'Status', type: 'status', path: 'status.phase' },
      {
        id: 'READY',
        title: 'Ready',
        type: 'string',
        // calculate a cell's value when it's not a simple, direct lookup
        accessor: (obj: any) => {
          const statuses = (obj?.status?.containerStatuses as any[]) ?? [];
          const total = statuses.length;
          const ready = statuses.filter((s) => s?.ready).length;
          return `${ready}/${total}`;
        },
      },
      { id: 'AGE', title: 'Age', type: 'date', path: 'metadata.creationTimestamp' },
      { id: 'IP', title: 'IP', type: 'ip', path: 'status.podIP' },
      { id: 'NODE', title: 'Node', type: 'string', path: 'spec.nodeName' },
    ],
  },
};


