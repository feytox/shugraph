import {
  skipToken,
  useMutation,
  useMutationState,
  useQuery,
  useQueryClient,
  type InfiniteData,
  type MutationState,
} from '@tanstack/react-query';
import {
  apiClient,
  type Group,
  type GroupListResponse,
  type GroupSolutionInput,
} from '@/api/client';
import { groupDetailKey, groupsListKey } from '@/shared/groupQueries';

const saveError = 'Не удалось сохранить решение';

class GroupChanged extends Error {}

type FailedDecision = { comment: string; message: string };

export function useGroupDecision(groupId: string) {
  const queryClient = useQueryClient();
  const mutationKey = ['group-decision', groupId];
  const failedDecisionKey = [...mutationKey, 'failure'];
  const { data: failedDecision } = useQuery<FailedDecision | null>({
    queryKey: failedDecisionKey,
    queryFn: skipToken,
    gcTime: Infinity,
  });
  const groupDecision = { mutationKey, exact: true };
  const detail = { queryKey: groupDetailKey(groupId) };
  const list = { queryKey: groupsListKey, exact: true };
  const latest = useMutationState<MutationState<unknown, Error, GroupSolutionInput>>({
    filters: groupDecision,
    select: ({ state }) => state,
  }).at(-1);
  const pending = latest?.status === 'pending';

  const refreshGroup = () =>
    Promise.all([queryClient.invalidateQueries(detail), queryClient.invalidateQueries(list)]);

  const mutation = useMutation({
    mutationKey,
    mutationFn: async (input: GroupSolutionInput) => {
      const { data, response } = await apiClient
        .PATCH('/api/v1/groups/{group_id}/solution', {
          params: { path: { group_id: groupId } },
          body: input,
        })
        .catch(() => {
          throw new Error(saveError);
        });
      if (response.status === 409) {
        throw new GroupChanged('По группе уже вынесено решение');
      }
      if (response.status === 404) throw new GroupChanged('Группа не найдена');
      if (!response.ok || !data) throw new Error(saveError);
      return data;
    },
    onSuccess: async (solution) => {
      queryClient.setQueryData(failedDecisionKey, null);
      await Promise.all([queryClient.cancelQueries(detail), queryClient.cancelQueries(list)]);
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
      await refreshGroup();
    },
    onError: async (error, input) => {
      queryClient.setQueryData(failedDecisionKey, {
        comment: input.decision_comment,
        message: error.message,
      } satisfies FailedDecision);
      if (error instanceof GroupChanged) await refreshGroup();
    },
  });

  const submittedComment = pending ? latest.variables?.decision_comment : failedDecision?.comment;

  return {
    save(input: GroupSolutionInput) {
      if (!queryClient.isMutating(groupDecision)) mutation.mutate(input);
    },
    pending,
    error: pending ? undefined : failedDecision?.message,
    submittedComment,
  };
}
