import { replaceEqualDeep, useInfiniteQuery, type InfiniteData } from '@tanstack/react-query';
import { apiClient, type GroupListResponse } from '@/api/client';
import { groupsListKey } from '@/shared/groupQueries';

const pollIntervalMs = 3000;

type GroupsQueryState = {
  status: string;
  fetchMeta: { fetchMore?: { direction: 'forward' | 'backward' } } | null;
};

function failedOnNextPage(state: GroupsQueryState) {
  return state.status === 'error' && state.fetchMeta?.fetchMore?.direction === 'forward';
}

function isGroupsPages(data: unknown): data is InfiniteData<GroupListResponse> {
  return (
    typeof data === 'object' &&
    data !== null &&
    'pages' in data &&
    Array.isArray(data.pages) &&
    'pageParams' in data &&
    Array.isArray(data.pageParams)
  );
}

function keepSavedDecisions(previous: unknown, incoming: unknown) {
  // Query applies structural sharing to both the paginated data and the selected list.
  if (!isGroupsPages(previous) || !isGroupsPages(incoming)) {
    return replaceEqualDeep(previous, incoming);
  }
  const decided = new Map(
    previous.pages
      .flatMap((page) => page.items)
      .filter((item) => item.status !== 'new')
      .map(({ id, status }) => [id, status])
  );
  if (!decided.size) return replaceEqualDeep(previous, incoming);

  // A request that started before a decision can still contain the earlier status.
  return replaceEqualDeep(previous, {
    ...incoming,
    pages: incoming.pages.map((page) => ({
      ...page,
      items: page.items.map((item) => {
        const saved = decided.get(item.id);
        return saved && item.status === 'new' ? { ...item, status: saved } : item;
      }),
    })),
  });
}

export function useGroups() {
  return useInfiniteQuery({
    queryKey: groupsListKey,
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam, signal }) => {
      const { data, response } = await apiClient.GET('/api/v1/groups', {
        params: { query: { limit: 20, cursor: pageParam } },
        signal,
      });
      if (!response.ok || !data) throw new Error('Не удалось загрузить список групп');
      return data;
    },
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
    structuralSharing: keepSavedDecisions,
    select: (data) => ({
      items: [
        ...new Map(
          data.pages.flatMap((page) => page.items).map((item) => [item.id, item])
        ).values(),
      ],
    }),
    retry: false,
    staleTime: pollIntervalMs,
    refetchInterval: ({ state }) =>
      state.data && !failedOnNextPage(state) ? pollIntervalMs : false,
    refetchOnWindowFocus: ({ state }) => (failedOnNextPage(state) ? false : 'always'),
  });
}
