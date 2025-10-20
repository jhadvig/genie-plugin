import * as React from 'react';
import { useEffect, useState } from 'react';
import { Table, Thead, Tr, Th, Tbody, Td } from '@patternfly/react-table';
import { Button, Modal, ModalVariant, Form, FormGroup, TextInput, Alert, Spinner, ModalHeader } from '@patternfly/react-core';
import { DashboardMCPClient } from '../services/dashboardClient';
import {GenieLayout } from './shared';

export default function GenieLibraryPage() {
  const [dashboards, setDashboards] = useState<DashboardListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isRenameOpen, setIsRenameOpen] = useState(false);
  const [renameTarget, setRenameTarget] = useState<DashboardListItem | null>(null);
  const [draftName, setDraftName] = useState('');
  const [draftDesc, setDraftDesc] = useState('');
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);


  useEffect(() => {
    const client = new DashboardMCPClient(
      `${window.location.origin}/api/proxy/plugin/genie-plugin/dashboard-mcp/`,
    );
    client
      .listDashboards()
      .then(({ layouts }) => {
        setDashboards(layouts || []);
        setLoading(false);
      })
      .catch((e) => {
        setError(e?.message || 'Failed to load dashboards');
        setLoading(false);
      });
  }, []);

  return (
    <GenieLayout title="Library">
    <div style={{ padding: '20px' }}>
      {!loading && !error && (
        <Table aria-label="Dashboards table" variant="compact">
          <Thead>
            <Tr>
              <Th>Name</Th>
              <Th>Description</Th>
              <Th>Active</Th>
              <Th>Created</Th>
              <Th>Updated</Th>
              <Th>Layout ID</Th>
              <Th aria-label="Actions">Actions</Th>
            </Tr>
          </Thead>
          <Tbody>
            {dashboards.map((d) => (
              <Tr key={d.id}>
                <Td modifier="nowrap">
                  <a href={`/genie/widgets?dashboardId=${d.layoutId}`}>
                    {d.name || d.layoutId}
                  </a>
                </Td>
                <Td modifier="nowrap">{d.description || '-'}</Td>
                <Td>{d.isActive ? '✓' : ''}</Td>
                <Td>{d.createdAt ? new Date(d.createdAt).toLocaleString() : '-'}</Td>
                <Td>{d.updatedAt ? new Date(d.updatedAt).toLocaleString() : '-'}</Td>
                <Td ><code style={{ fontSize: '0.85em' }}>{d.layoutId}</code></Td>
                <Td>
                  <Button
                    variant="secondary"
                    onClick={() => {
                      setRenameTarget(d);
                      setDraftName(d.name || '');
                      setDraftDesc(d.description || '');
                      setSaveError(null);
                      setIsRenameOpen(true);
                    }}
                  >
                    Edit
                  </Button>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>
      )}

      <Modal
        variant={ModalVariant.small}
        title="Rename dashboard"
        isOpen={isRenameOpen}
        onClose={() => setIsRenameOpen(false)}
        style={{ padding: '20px' }}
      >
        {saveError && (
          <Alert isInline variant="danger" title="Error updating dashboard">
            {saveError}
          </Alert>
        )}
        <ModalHeader title="Edit Dashboard" labelId="edit-dashboard" />
        <Form>
          <FormGroup label="Name" isRequired fieldId="rename-name">
            <TextInput id="rename-name" value={draftName} onChange={(_, newVal) => setDraftName(newVal)} isRequired />
          </FormGroup>
          <FormGroup label="Description" fieldId="rename-desc">
            <TextInput id="rename-desc" value={draftDesc} onChange={(_, newVal) => setDraftDesc(newVal)} />
          </FormGroup>
        </Form>
        <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
          <Button
            variant="primary"
            isDisabled={saving || !draftName.trim()}
            onClick={async () => {
              if (!renameTarget) return;
              setSaving(true);
              setSaveError(null);
              try {
                const client = new DashboardMCPClient(
                  `${window.location.origin}/api/proxy/plugin/genie-plugin/dashboard-mcp/`,
                );
                await client.setDashboardMetadata(renameTarget.layoutId, draftName.trim(), draftDesc || undefined);
                setDashboards((prev) => prev.map((row) => row.id === renameTarget.id ? { ...row, name: draftName.trim(), description: draftDesc } : row));
                setIsRenameOpen(false);
              } catch (e: any) {
                setSaveError(e?.message || 'Failed to save changes');
              } finally {
                setSaving(false);
              }
            }}
          >
            {saving ? <><Spinner size="sm" /> Saving</> : 'Save'}
          </Button>
          <Button variant="link" onClick={() => setIsRenameOpen(false)}>Cancel</Button>
        </div>
      </Modal>
    </div>
    </GenieLayout>
  );
}

export type DashboardListItem = {
  id: string;
  layoutId: string;
  name: string;
  description: string;
  isActive?: boolean;
  createdAt?: string;
  updatedAt?: string;
};
