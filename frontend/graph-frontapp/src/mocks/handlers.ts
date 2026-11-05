import { delay, http, HttpResponse } from 'msw';
import type { GroupItem, GroupListResponse } from '@/api/client';

const demoGroup: GroupItem = {
  id: '00000000-0000-4000-8000-000000000001',
  scheme: 'farm',
  status: 'new',
  risk_level: 'high',
  risk_score: 87,
  risk_parts: {},
  detected_at: '2026-10-07T10:00:00Z',
  first_event_at: '2026-10-07T10:00:00Z',
  last_event_at: '2026-10-07T10:30:45Z',
  features: { accounts_count: 7, total_amount: 12345, span_minutes: 30 },
};

export const handlers = [
  http.get('*/api/v1/groups', async () => {
    switch (import.meta.env.VITE_GROUPS_MOCK_STATE) {
      case 'loading':
        await delay('infinite');
        break;
      case 'error':
        return new HttpResponse(null, { status: 500 });
      case 'empty':
        return HttpResponse.json<GroupListResponse>({ items: [], next_cursor: null });
    }
    return HttpResponse.json<GroupListResponse>({ items: [demoGroup], next_cursor: null });
  }),
];
