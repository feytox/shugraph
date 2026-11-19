import {
  useIsMutating,
  useMutation,
  useQueryClient,
  type InfiniteData,
} from '@tanstack/react-query';
import {
  apiClient,
  type Group,
  type GroupListResponse,
  type GroupSolutionInput,
} from '@/api/client';
import { groupDetailKey, groupsListKey } from '@/shared/groupQueries';

const saveError = 'Не удалось сохранить решение';

class DecisionConflict extends Error {}

export function useGroupDecision(groupId: string) {
  const queryClient = useQueryClient();
  const mutationKey = ['group-decision', groupId];
  const groupDecision = { mutationKey, exact: true };
  const detail = { queryKey: groupDetailKey(groupId), exact: true };
  const list = { queryKey: groupsListKey, exact: true };
  const pending = useIsMutating(groupDecision) > 0;

  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ ...detail, refetchType: 'all' }),
      queryClient.invalidateQueries(list),
    ]);

  const mutation = useMutation({
    mutationKey,
    mutationFn: async (input: GroupSolutionInput) => {
      const { data, response } = await apiClient.PATCH('/api/v1/groups/{group_id}/solution', {
        params: { path: { group_id: groupId } },
        body: input,
      });
      if (response.status === 409) throw new DecisionConflict('Не удалось обновить статус группы');
      if (!response.ok || !data) throw new Error(saveError);
      return data;
    },
    onSuccess: async (solution) => {
      await Promise.all([queryClient.cancelQueries(detail), queryClient.cancelQueries(list)]);
      queryClient.setQueryData<Group>(detail.queryKey, (group) =>
        group ? { ...group, ...solution } : group
      );
      queryClient.setQueryData<InfiniteData<GroupListResponse>>(
        list.queryKey,
        (groups) =>
          groups && {
            ...groups,
            pages: groups.pages.map((page) => ({
              ...page,
              items: page.items.map((item) =>
                item.id === groupId ? { ...item, status: solution.status } : item
              ),
            })),
          }
      );
      await refresh();
    },
    onError: (error) => (error instanceof DecisionConflict ? refresh() : undefined),
  });

  return {
    save(input: GroupSolutionInput) {
      if (!queryClient.isMutating(groupDecision)) mutation.mutate(input);
    },
    pending,
    error: mutation.isError
      ? mutation.error instanceof DecisionConflict
        ? mutation.error.message
        : saveError
      : null,
  };
}
