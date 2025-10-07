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

class DashboardUtils {
  static normalizeResponse(
    dashboard: CreateDashboardResponse | ActiveDashboardResponse | undefined,
  ): NormalizedDashboard | undefined {
    if (!dashboard) return undefined;

    const maybeActive = dashboard as ActiveDashboardResponse;
    if (maybeActive.analysis && maybeActive.activeLayoutId) {
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

    return widgets
      .filter((w) => w && (w as any).id != null)
      .map((w) => {
        const defaultPos = { x: 0, y: 0, w: 4, h: 4 };
        const pos = (w as any).position ?? defaultPos;
        const safeNumber = (value: unknown, fallback: number) => {
          const num = typeof value === 'number' ? value : Number(value);
          return Number.isFinite(num) ? num : fallback;
        };

        const normalizedPosition = {
          x: safeNumber((pos as any).x, defaultPos.x),
          y: safeNumber((pos as any).y, defaultPos.y),
          w: Math.max(1, safeNumber((pos as any).w, defaultPos.w)),
          h: Math.max(1, safeNumber((pos as any).h, defaultPos.h)),
        };

        return {
          id: String((w as any).id),
          componentType: (w as any).componentType ?? 'chart',
          position: normalizedPosition,
          props: (w as any).props ?? {},
          breakpoint: (w as any).breakpoint ?? 'lg',
        } as DashboardWidget;
      });
  }
}

export default DashboardUtils;