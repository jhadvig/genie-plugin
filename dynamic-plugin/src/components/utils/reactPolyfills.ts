/* eslint-disable @typescript-eslint/ban-ts-comment */
import { useMemo, useLayoutEffect } from 'react';
import * as React from 'react';

export async function patchUseId() {
  // @ts-ignore
  const scope = __webpack_share_scopes__?.default;
  if (!scope) {
    return;
  }
  if (scope) {
    let react = await scope.react['*'].get();
    if (!react) {
      return;
    }
    react = react();
    if (!react.useId) {
      console.log('[Genie] Patching React.useId for compatibility');
      react.useId = () => {
        const id = useMemo(() => {
          return crypto.randomUUID();
        }, []);
        return id;
      };
    }
  }
}

// Polyfill for React 17 compatibility with libraries expecting React 18
export function setupReactPolyfills() {
  if (!(React as any).useInsertionEffect) {
    (React as any).useInsertionEffect = useLayoutEffect;
  }
}

// Initialize polyfills
patchUseId();
setupReactPolyfills();