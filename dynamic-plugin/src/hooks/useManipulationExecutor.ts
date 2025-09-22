import { useState, useEffect, useCallback } from 'react';
import { ManipulateWidgetResponse } from '../types/dashboard';

interface PendingManipulation {
  id: string;
  manipulation: ManipulateWidgetResponse;
  timestamp: number;
  executed: boolean;
}

export function useManipulationExecutor() {
  const [pendingManipulations, setPendingManipulations] = useState<PendingManipulation[]>([]);
  const [executedManipulations, setExecutedManipulations] = useState<string[]>([]);

  // Add new manipulations to the pending list
  const addManipulation = useCallback((manipulation: ManipulateWidgetResponse) => {
    const id = `${manipulation.layoutId}-${manipulation.timestamp}`;

    setPendingManipulations((prev) => {
      // Check if this manipulation already exists
      const exists = prev.some((m) => m.id === id);
      if (exists) {
        return prev;
      }

      return [
        ...prev,
        {
          id,
          manipulation,
          timestamp: Date.now(),
          executed: false,
        },
      ];
    });
  }, []);

  // Mark a manipulation as executed and remove it from pending
  const executeManipulation = useCallback((id: string) => {
    setPendingManipulations((prev) => prev.filter((m) => m.id !== id));
    setExecutedManipulations((prev) => [...prev, id]);
  }, []);

  // Get the next manipulation to execute
  const getNextManipulation = useCallback(() => {
    return pendingManipulations.find((m) => !m.executed) || null;
  }, [pendingManipulations]);

  // Clean up old executed manipulations (keep only last 10)
  useEffect(() => {
    setExecutedManipulations((prev) => prev.slice(-10));
  }, [JSON.stringify(executedManipulations)]);

  return {
    pendingManipulations,
    addManipulation,
    executeManipulation,
    getNextManipulation,
    hasPendingManipulations: pendingManipulations.length > 0,
    executedCount: executedManipulations.length,
  };
}
