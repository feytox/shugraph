import { useState } from 'react';
import { focusManager, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient, type GroupListResponse } from '@/api/client';

const queryKey = ['groups'];

async function page(cursor?: string, signal?: AbortSignal, limit = 20) {
  const { data, error, response } = await apiClient.GET('/api/v1/groups', {
    params: { query: { limit, cursor } },
    signal,
  });
  if (!response.ok || error !== undefined || !data) {
    throw new Error('Не удалось загрузить список групп');
  }
  return data;
}

export function useGroups() {
  const client = useQueryClient();
  const [isRefetchError, setRefetchError] = useState(false);
  const next = useMutation({
    onMutate: () => {
      const refresh = { needed: false };
      const unsubscribe = focusManager.subscribe((focused) => {
        if (focused) refresh.needed = true;
      });
      return { refresh, unsubscribe };
    },
    mutationFn: async () => {
      // A focus refresh may already be in flight. Its stale response must not replace this page.
      await client.cancelQueries({ queryKey });
      const previous = client.getQueryData<GroupListResponse>(queryKey)!;
      return page(previous.next_cursor ?? undefined);
    },
    onSuccess: (data) => {
      client.setQueryData<GroupListResponse>(queryKey, (previous) => ({
        items: [
          ...new Map(
            [...(previous?.items ?? []), ...data.items].map((item) => [item.id, item])
          ).values(),
        ],
        next_cursor: data.next_cursor,
      }));
    },
    onSettled: (_data, _error, _variables, context) => {
      context?.unsubscribe();
      if (context?.refresh.needed) {
        // Re-enabling the query must refresh even after pagination marks the cache fresh.
        void client.invalidateQueries({ queryKey, refetchType: 'none' });
      }
    },
  });
  const query = useQuery({
    queryKey,
    queryFn: async ({ signal }) => {
      const previous = client.getQueryData<GroupListResponse>(queryKey);
      try {
        const remaining = new Set(previous?.items.map((item) => item.id));
        const items = new Map<string, GroupListResponse['items'][number]>();
        let cursor: string | undefined;
        do {
          const data = await page(
            cursor,
            signal,
            previous?.items.length ? Math.min(20, Math.max(1, remaining.size)) : 20
          );
          for (const item of data.items) {
            items.set(item.id, item);
            remaining.delete(item.id);
          }
          cursor = data.next_cursor ?? undefined;
        } while (cursor && remaining.size);
        signal.throwIfAborted();
        setRefetchError(false);
        return { items: [...items.values()], next_cursor: cursor ?? null };
      } catch (error) {
        if (previous && !signal.aborted) setRefetchError(true);
        throw error;
      }
    },
    // Serialize pagination and refreshes, including focus changes during a next-page request.
    enabled: !next.isPending,
    retry: false,
    staleTime: 3000,
    refetchInterval: 3000,
    refetchOnWindowFocus: 'always',
  });
  return {
    ...query,
    isRefetchError,
    isFetching: query.isFetching || next.isPending,
    isFetchingNextPage: next.isPending,
    isFetchNextPageError: next.isError,
    hasNextPage: Boolean(query.data?.next_cursor),
    fetchNextPage: () => next.mutate(),
  };
}
