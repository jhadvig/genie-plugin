import { AbsoluteTimeRange, RelativeTimeRange, TimeRangeValue } from '@perses-dev/core';
import { useMemo } from 'react';

export const useTimeRange = (start?: string, end?: string, duration?: string) => {
  const result = useMemo(() => {
    if (end === 'NOW') {
      // Special placeholder for current-time queries.
      end = undefined;
    }
    let timeRange: TimeRangeValue;
    if (start && end) {
      timeRange = {
        start: new Date(start),
        end: new Date(end),
      } as AbsoluteTimeRange;
    } else {
      timeRange = { pastDuration: duration || '1h' } as RelativeTimeRange;
    }
    return timeRange;
  }, [duration, end, start]);
  return result;
};
