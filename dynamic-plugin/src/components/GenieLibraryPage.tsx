import * as React from 'react';
import { useEffect, useState } from 'react';
import { Gallery, GalleryItem, Card, CardTitle, CardBody, CardHeader } from '@patternfly/react-core';
import { DashboardMCPClient } from '../services/dashboardClient';
import {GenieLayout } from './shared';

export default function GenieLibraryPage() {
  const [dashboards, setDashboards] = useState<DashboardListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const client = new DashboardMCPClient('http://localhost:9081/mcp');
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
        <Gallery hasGutter>
          {dashboards.map((d) => {
            return (
              <GalleryItem key={d.id}>
                <Card isCompact>
                  <CardHeader>
                    <CardTitle>{d.name || d.layoutId}</CardTitle>
                  </CardHeader>
                  <CardBody>
                    {d.description && (
                      <div>{d.description}</div>
                    )}
                  </CardBody>
                </Card>
              </GalleryItem>
            );
          })}
          {/* add error state */}
          {/* add loading state */}
          {/* add no dashboards state */}
        </Gallery>
      )}
    </div>
    </GenieLayout>
  );
}

export type DashboardListItem = {
    id: string;
    layoutId: string;
    name: string;
    description: string;
  };


