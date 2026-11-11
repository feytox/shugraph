import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';

export function useGroupGraph(groupId: string) {
  return useQuery({
    queryKey: ['groups', groupId, 'graph'],
    queryFn: async ({ signal }) => {
      const { data, response } = await apiClient.GET('/api/v1/groups/{group_id}/graph', {
        params: { path: { group_id: groupId } },
        signal,
      });
      if (response.status === 404) return null;
      if (!response.ok || !data) throw new Error('Не удалось загрузить граф группы');
      return data;
    },
    retry: false,
  });
}
