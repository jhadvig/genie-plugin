import { ActiveDashboardResponse, CreateDashboardResponse, DashboardWidget } from "../../types/dashboard";

export type NormalizedDashboard = {
  activeLayoutId: string;
  layout: {
    layoutId: string;
    name: string;
    description: string;
  };
  widgets: DashboardWidget[];
};

type RawWidget = Partial<DashboardWidget> & {
  i?: string;
  id?: string;
  x?: number;
  y?: number;
  w?: number;
  h?: number;
  position?: { x?: number; y?: number; w?: number; h?: number };
  persesComponent?: string;
};

class DashboardUtils {
  static normalizeResponse(
    dashboard: CreateDashboardResponse | ActiveDashboardResponse | undefined,
  ): NormalizedDashboard | undefined {
    if (!dashboard) return undefined;

    const maybeActive = dashboard as ActiveDashboardResponse;
    if (maybeActive.analysis && maybeActive.activeLayoutId) {
      console.log('maybeActive', maybeActive);
      return {
        activeLayoutId: maybeActive.activeLayoutId,
        layout: {
          layoutId: maybeActive.analysis.layoutId,
          name: maybeActive.analysis.name,
          description: maybeActive.analysis.description,
        },
        widgets: DashboardUtils.normalizeWidgets(maybeActive.analysis.widgets),
      };
    }

    const created = dashboard as CreateDashboardResponse;

    console.log({ created });
    return {
      activeLayoutId: created.activeLayoutId,
      layout: {
        layoutId: created.layout?.layoutId ?? created.activeLayoutId,
        name: created.layout?.name ?? 'Untitled Dashboard',
        description: created.layout?.description ?? created.message ?? '',
      },
      widgets: DashboardUtils.normalizeWidgets(created.widgets),
    };
  }

  static normalizeWidgets(widgets?: DashboardWidget[]): DashboardWidget[] {
    if (!widgets || widgets.length === 0) return [];

    const safeNumber = (value: unknown, fallback: number): number => {
      const num = typeof value === 'number' ? value : Number(value);
      return Number.isFinite(num) ? num : fallback;
    };

    return (widgets as RawWidget[])
      .filter((w): w is RawWidget => w != null && (w.id != null || w.i != null))
      .map((w): DashboardWidget => {
        const defaultPos = { x: 0, y: 0, w: 4, h: 4 };
        const pos = w.position ?? w;

        const normalizedPosition = {
          x: safeNumber(pos.x, defaultPos.x),
          y: safeNumber(pos.y, defaultPos.y),
          w: Math.max(1, safeNumber(pos.w, defaultPos.w)),
          h: Math.max(1, safeNumber(pos.h, defaultPos.h)),
        };

        const componentType = w.componentType ?? 'chart';
        const existingProps = w.props ?? {};
        const persesComponent = w.persesComponent ?? existingProps.persesComponent ?? (componentType === 'chart' ? 'PersesTimeSeries' : undefined);

        return {
          id: String(w.id ?? w.i),
          componentType,
          position: normalizedPosition,
          props: {
            ...existingProps,
            ...(persesComponent && { persesComponent }),
          },
          breakpoint: w.breakpoint ?? 'lg',
        };
      });
  }

  static applyDashboardMetadataUpdate(
    setActiveDashboard: (updater: (prev: NormalizedDashboard | undefined) => NormalizedDashboard | undefined) => void,
    meta: { layoutId?: string; name?: string; description?: string },
  ) {
    setActiveDashboard((prev) => {
      if (!prev) return prev;
      if (meta.layoutId && prev.layout.layoutId !== meta.layoutId) return prev;
      return {
        ...prev,
        layout: {
          ...prev.layout,
          ...(meta.name !== undefined ? { name: meta.name } : {}),
          ...(meta.description !== undefined ? { description: meta.description } : {}),
        },
      };
    });
  }
}

export default DashboardUtils;