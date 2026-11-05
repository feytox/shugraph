import { useInfiniteQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';

export function useGroups() {
  const query = useInfiniteQuery({
    queryKey: ['groups'],
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam, signal }) => {
      const { data, error, response } = await apiClient.GET('/api/v1/groups', {
        params: { query: { limit: 20, cursor: pageParam } },
        signal,
      });
      if (!response.ok || error !== undefined || !data) {
        throw new Error('Не удалось загрузить список групп');
      }
      return data;
    },
    getNextPageParam: (page) => page.next_cursor ?? undefined,
    retry: false,
    refetchInterval: 3000,
    refetchOnWindowFocus: 'always',
  });
  return {
    ...query,
    data: query.data && {
      items: [
        ...new Map(
          query.data.pages.flatMap((page) => page.items).map((item) => [item.id, item])
        ).values(),
      ],
      next_cursor: query.data.pages.at(-1)?.next_cursor ?? null,
    },
  };
}
