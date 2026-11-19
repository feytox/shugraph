import {
  useMutation,
  useMutationState,
  useQueryClient,
  type InfiniteData,
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
  const latest = useMutationState({ filters: groupDecision, select: ({ state }) => state }).at(-1);
  const pending = latest?.status === 'pending';

  const pendingNextPage = () => {
    const query = queryClient.getQueryCache().find(list);
    if (
      query &&
      query.state.fetchStatus !== 'idle' &&
      query.state.fetchMeta?.fetchMore?.direction === 'forward'
    ) {
      return query.promise;
    }
  };

  const refreshList = (canRefetch = true) =>
    queryClient.invalidateQueries(
      { ...list, refetchType: canRefetch ? 'active' : 'none' },
      { cancelRefetch: false }
    );

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
      const nextPage = pendingNextPage();
      await Promise.all([
        queryClient.cancelQueries(detail),
        nextPage ? undefined : queryClient.cancelQueries(list),
      ]);
      queryClient.setQueriesData<Group>(detail, (group) =>
        group ? { ...group, ...solution, is_frozen: true } : group
      );
      const updateListStatus = () =>
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
      updateListStatus();

      const refreshDetails = queryClient.invalidateQueries(detail);
      const pageLoaded = await nextPage?.then(
        () => true,
        () => false
      );
      // The next page response can include the earlier snapshot of already loaded pages.
      if (pageLoaded) updateListStatus();
      await Promise.all([refreshDetails, refreshList(pageLoaded !== false)]);
    },
    onError: async (error) => {
      if (!(error instanceof DecisionConflict)) return;
      const refreshDetails = queryClient.invalidateQueries(detail);
      await pendingNextPage()?.catch(() => undefined);
      await Promise.all([refreshDetails, refreshList()]);
    },
  });

  const submitted = latest?.variables;
  const submittedComment =
    (pending || latest?.status === 'error') &&
    submitted &&
    typeof submitted === 'object' &&
    'decision_comment' in submitted &&
    typeof submitted.decision_comment === 'string'
      ? submitted.decision_comment
      : undefined;

  return {
    save(input: GroupSolutionInput) {
      if (!queryClient.isMutating(groupDecision)) mutation.mutate(input);
    },
    pending,
    error: latest?.error?.message,
    submittedComment,
  };
}
