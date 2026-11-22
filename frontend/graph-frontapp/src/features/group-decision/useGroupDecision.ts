import {
  useMutation,
  useMutationState,
  useQueryClient,
  type InfiniteData,
  type MutationState,
} from '@tanstack/react-query';
import {
  apiClient,
  getProblemDetail,
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
  const detail = { queryKey: groupDetailKey(groupId) };
  const list = { queryKey: groupsListKey, exact: true };
  const latest = useMutationState<MutationState<unknown, Error, GroupSolutionInput>>({
    filters: groupDecision,
    select: ({ state }) => state,
  }).at(-1);
  const pending = latest?.status === 'pending';

  const refreshList = () => queryClient.invalidateQueries(list, { cancelRefetch: false });

  const mutation = useMutation({
    mutationKey,
    mutationFn: async (input: GroupSolutionInput) => {
      const { data, error, response } = await apiClient
        .PATCH('/api/v1/groups/{group_id}/solution', {
          params: { path: { group_id: groupId } },
          body: input,
        })
        .catch(() => {
          throw new Error(saveError);
        });
      const reason = getProblemDetail(error);
      if (response.status === 409) {
        throw new DecisionConflict(reason ?? 'Не удалось обновить статус группы');
      }
      if (!response.ok || !data) throw new Error(reason ?? saveError);
      return data;
    },
    onSuccess: async (solution) => {
      await queryClient.cancelQueries(detail);
      queryClient.setQueriesData<Group>(detail, (group) =>
        group ? { ...group, ...solution, is_frozen: true } : group
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
      await Promise.all([queryClient.invalidateQueries(detail), refreshList()]);
    },
    onError: async (error) => {
      if (!(error instanceof DecisionConflict)) return;
      const listWasFetching = queryClient.isFetching(list) > 0;
      const refreshListAfterConflict = async () => {
        await refreshList();
        if (listWasFetching) await refreshList();
      };
      await Promise.all([queryClient.invalidateQueries(detail), refreshListAfterConflict()]);
    },
  });

  const submittedComment =
    pending || latest?.status === 'error' ? latest?.variables?.decision_comment : undefined;

  return {
    save(input: GroupSolutionInput) {
      if (!queryClient.isMutating(groupDecision)) mutation.mutate(input);
    },
    pending,
    error: latest?.error?.message,
    submittedComment,
  };
}
