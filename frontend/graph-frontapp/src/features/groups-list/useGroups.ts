import { useInfiniteQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';

export function useGroups() {
  return useInfiniteQuery({
    queryKey: ['groups'],
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam, signal }) => {
      const { data, error, response } = await apiClient.GET('/api/v1/groups', {
        params: { query: { limit: 20, cursor: pageParam } },
        signal,
      });
      if (!response.ok || error !== undefined || !data) {
        throw new Error(error?.detail || 'Не удалось загрузить список групп');
      }
      return data;
    },
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
    select: (data) => ({
      items: [
        ...new Map(
          data.pages.flatMap((page) => page.items).map((item) => [item.id, item])
        ).values(),
      ],
    }),
    retry: false,
    staleTime: 3000,
    refetchInterval: (query) => (query.state.data ? 3000 : false),
    refetchOnWindowFocus: 'always',
  });
}
