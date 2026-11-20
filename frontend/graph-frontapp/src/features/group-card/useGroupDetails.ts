import { useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient, getProblemDetail, type Group, type GroupItem } from '@/api/client';
import { groupDetailKey } from '@/shared/groupQueries';

const loadError = 'Не удалось загрузить карточку группы';

export class GroupNotFound extends Error {}

export function useGroupDetails(summary: GroupItem) {
  const queryClient = useQueryClient();
  const groupId = summary.id;
  const detailKey = groupDetailKey(groupId);

  return useQuery({
    queryKey: [...detailKey, summary],
    initialData: () =>
      queryClient
        .getQueriesData<Group>({ queryKey: detailKey })
        .reverse()
        .find(([, group]) => group !== undefined)?.[1],
    initialDataUpdatedAt: 0,
    queryFn: async ({ signal }) => {
      const { data, error, response } = await apiClient
        .GET('/api/v1/groups/{group_id}', {
          params: { path: { group_id: groupId } },
          signal,
        })
        .catch(() => {
          throw new Error(loadError);
        });
      const reason = getProblemDetail(error);
      if (response.status === 404) {
        throw new GroupNotFound(reason ?? 'Карточка группы не найдена');
      }
      if (!response.ok || !data) throw new Error(reason ?? loadError);
      return data;
    },
    retry: false,
  });
}
