import { http, HttpResponse } from 'msw';
import type {
  ErrorModel,
  Group,
  GroupGraph,
  GroupListResponse,
  GroupSolutionInput,
  GroupSolutionResponse,
} from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';
import { createGroupsDemo } from '@/mocks/groupsDemo';
import { groupsPage } from '@/mocks/groupsPage';

export function createHandlers() {
  const demo = createGroupsDemo();
  return [
    http.get('*/api/v1/groups', ({ request }) =>
      HttpResponse.json<GroupListResponse>(groupsPage(demo.getItems(), request))
    ),

    http.patch<{ groupId: string }, GroupSolutionInput, GroupSolutionResponse | ErrorModel>(
      '*/api/v1/groups/:groupId/solution',
      async ({ params: { groupId }, request }) => {
        const input = parseSolution(await request.json().catch(() => undefined));
        if (!input)
          return problem(400, 'Bad Request', 'a status and a decision comment are required');
        const group = demo.getItem(groupId);
        if (!group) return groupNotFound();
        if (group.status !== 'new') return problem(409, 'Conflict', 'group already has a decision');
        return HttpResponse.json<GroupSolutionResponse>(demo.saveSolution(group, input));
      }
    ),

    http.get<{ groupId: string }, never, Group | ErrorModel>(
      '*/api/v1/groups/:groupId',
      ({ params }) => {
        const group = demo.getGroup(params.groupId);
        if (!group) return groupNotFound();
        return HttpResponse.json<Group>(group);
      }
    ),

    http.get<{ groupId: string }, never, GroupGraph | ErrorModel>(
      '*/api/v1/groups/:groupId/graph',
      ({ params }) => {
        const group = mockGroups.find((mock) => mock.item.id === params.groupId);
        if (!group) return groupNotFound();
        return HttpResponse.json<GroupGraph>(group.graph);
      }
    ),
  ];
}

function problem(status: number, title: string, detail: string) {
  return HttpResponse.json<ErrorModel>(
    { type: 'about:blank', title, status, detail },
    { status, headers: { 'Content-Type': 'application/problem+json' } }
  );
}

const groupNotFound = () => problem(404, 'Not Found', 'group not found');

function parseSolution(input: unknown): GroupSolutionInput | undefined {
  if (
    !input ||
    typeof input !== 'object' ||
    !('status' in input) ||
    (input.status !== 'approved' && input.status !== 'declined') ||
    !('decision_comment' in input) ||
    typeof input.decision_comment !== 'string' ||
    !input.decision_comment.trim()
  ) {
    return undefined;
  }
  return { status: input.status, decision_comment: input.decision_comment };
}
