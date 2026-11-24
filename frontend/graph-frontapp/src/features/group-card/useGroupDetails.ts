import { useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient, type Group, type GroupItem } from '@/api/client';
import { groupDetailKey, groupsPollIntervalMs } from '@/shared/groupQueries';

const loadError = 'Не удалось загрузить карточку группы';

export class GroupNotFound extends Error {}

export function useGroupDetails(groupId: string, summary: GroupItem | undefined) {
  const queryClient = useQueryClient();
  const detailKey = groupDetailKey(groupId);

  return useQuery({
    queryKey: [
      ...detailKey,
      summary?.last_event_at,
      summary?.status,
      summary?.risk_score,
      summary?.risk_level,
      summary?.features,
      summary?.risk_parts,
    ],
    initialData: () =>
      queryClient
        .getQueriesData<Group>({ queryKey: detailKey })
        .reverse()
        .find(([, group]) => group !== undefined)?.[1],
    // The cached copy keeps the draft visible while the current details load.
    initialDataUpdatedAt: 0,
    queryFn: async ({ signal }) => {
      const { data, response } = await apiClient
        .GET('/api/v1/groups/{group_id}', {
          params: { path: { group_id: groupId } },
          signal,
        })
        .catch(() => {
          throw new Error(loadError);
        });
      if (response.status === 404) {
        throw new GroupNotFound('Карточка группы не найдена');
      }
      if (!response.ok || !data) throw new Error(loadError);
      return data;
    },
    retry: false,
    refetchInterval: ({ state }) => (!summary && state.data ? groupsPollIntervalMs : false),
  });
}
