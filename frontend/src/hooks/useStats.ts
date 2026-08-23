import { useQuery } from '@tanstack/react-query';
import { getStats } from '../api/client';

export function useStats() {
  return useQuery({
    queryKey: ['stats'],
    queryFn: getStats,
    // Poll quickly while links are being processed, otherwise every 30s
    refetchInterval: (query) => {
      const s = query.state.data;
      if (!s) return 30_000;
      const busy =
        (s.by_status['new'] ?? 0) +
        (s.by_status['fetching'] ?? 0) +
        (s.by_status['processing'] ?? 0);
      return busy > 0 ? 2_000 : 30_000;
    },
  });
}
