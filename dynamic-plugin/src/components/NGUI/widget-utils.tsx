import React from 'react';
import {
  K8sGroupVersionKind,
  K8sResourceCommon,
  ResourceLink,
  Timestamp,
  K8sModel,
} from '@openshift-console/dynamic-plugin-sdk';
import { ColumnSpec } from './dataTypeRegistry';

export type NGUIField = {
  id: string;
  name: string;
  data_path?: string;
};

const isModelWithSchema = (model: K8sModel): model is K8sModel & { openAPIV3Schema: any } => {
  return (model as any)?.openAPIV3Schema !== undefined;
};

// TODO: This is a temporary function to get a value by a path in a JSON object. We should use a more robust JSON path library.
// See https://github.com/rhamilto/console/blob/main/frontend/public/components/default-resource.tsx
export const getByPath = (obj: any, path?: string): any => {
  if (!obj || !path) return undefined;
  const parts = path.split('.');
  let current = obj;
  for (const p of parts) {
    if (current == null) return undefined;
    current = current[p];
  }
  return current;
};

export const matchFieldToColumnSpec = (field: NGUIField, allSpecs: ColumnSpec[]): ColumnSpec | undefined => {
  const byId = allSpecs.find((s) => s.id.toLowerCase() === field.id.toLowerCase());
  if (byId) return byId;
  const byName = allSpecs.find((s) => s.title.toLowerCase() === field.name.toLowerCase());
  return byName;
};

/**
 * Traverses the OpenAPI schema for a model to find the 'format' of a nested property.
 * @param model The K8sModel for the resource.
 * @param path A dot-notation path like 'metadata.creationTimestamp'.
 * @returns The format string (e.g., 'date-time') or undefined if not found.
 */
export const getSchemaFormat = (model?: K8sModel, path?: string): string | undefined => {
  if (!path || !model || !isModelWithSchema(model)) {
    return undefined;
  }

  const pathParts = path.split('.');
  let currentSchema = model.openAPIV3Schema;

  for (const part of pathParts) {
    const nextSchema = currentSchema?.properties?.[part];
    if (!nextSchema) {
      return undefined;
    }
    currentSchema = nextSchema;
  }

  return currentSchema?.format;
};

export const typeRenderers: Record<
  string,
  (value: any, obj: K8sResourceCommon, gvk: K8sGroupVersionKind) => React.ReactNode
> = {
  resource: (value, obj, gvk) => (
    <ResourceLink
      groupVersionKind={gvk}
      name={value as string}
      namespace={obj.metadata?.namespace}
    />
  ),
  'date-time': (value) => (value ? <Timestamp timestamp={value as string} /> : '-'),
};
