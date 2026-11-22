import { http, HttpResponse } from 'msw';
import type { ErrorModel, GroupGraph, GroupListResponse } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';
import { createGroupsDemo } from '@/mocks/groupsDemo';
import { groupsPage } from '@/mocks/groupsPage';

const demo = createGroupsDemo();

export const handlers = [
  http.get('*/api/v1/groups', ({ request }) =>
    HttpResponse.json<GroupListResponse>(groupsPage(demo.getItems(), request))
  ),

  http.get<{ groupId: string }, never, GroupGraph | ErrorModel>(
    '*/api/v1/groups/:groupId/graph',
    ({ params }) => {
      const group = mockGroups.find((mock) => mock.item.id === params.groupId);
      if (!group) {
        return HttpResponse.json<ErrorModel>(
          { type: 'about:blank', title: 'Not Found', status: 404, detail: 'group not found' },
          { status: 404, headers: { 'Content-Type': 'application/problem+json' } }
        );
      }
      return HttpResponse.json<GroupGraph>(group.graph);
    }
  ),
];
