import { delay, http, HttpResponse } from 'msw';
import type { GroupListResponse } from '@/api/client';
import { createGroupsDemo } from '@/mocks/groupsDemo';
import { groupsPage } from '@/mocks/groupsPage';

const demo = createGroupsDemo();

export const handlers = [
  http.get('*/api/v1/groups', async ({ request }) => {
    switch (import.meta.env.VITE_GROUPS_MOCK_STATE) {
      case 'loading':
        await delay('infinite');
        break;
      case 'error':
        return new HttpResponse(null, { status: 500 });
      case 'empty':
        return HttpResponse.json<GroupListResponse>({ items: [], next_cursor: null });
      case 'polling-error':
        if (demo.isRefreshError()) return new HttpResponse(null, { status: 500 });
        break;
      case 'next-page-error':
        if (new URL(request.url).searchParams.has('cursor')) {
          return new HttpResponse(null, { status: 500 });
        }
        break;
    }
    return HttpResponse.json<GroupListResponse>(groupsPage(demo.getItems(), request));
  }),
];
