import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { groupDetailKey } from '@/shared/groupQueries';

export function useGroupDetails(groupId: string) {
  return useQuery({
    queryKey: groupDetailKey(groupId),
    queryFn: async ({ signal }) => {
      const { data, response } = await apiClient.GET('/api/v1/groups/{group_id}', {
        params: { path: { group_id: groupId } },
        signal,
      });
      if (response.status === 404) return null;
      if (!response.ok || !data) throw new Error('Не удалось загрузить карточку группы');
      return data;
    },
    retry: false,
  });
}
