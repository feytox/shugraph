import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';

export function useGroups() {
  return useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET('/api/v1/groups', {
        params: { query: { limit: 20 } },
      });
      if (!response.ok || error !== undefined || !data) {
        throw new Error('Не удалось загрузить список групп');
      }
      return data;
    },
    retry: false,
  });
}
