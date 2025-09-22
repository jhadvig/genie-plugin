import { QueryClient } from '@tanstack/react-query';

const persesQueryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 0,
      queryFn: async ({ queryKey }) => {
        throw new Error(`No queryFn defined for queryKey: ${queryKey.join(',')}`);
      },
    },
  },
});

export default persesQueryClient;
