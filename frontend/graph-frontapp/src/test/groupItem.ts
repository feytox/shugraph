import type { GroupItem } from '@/api/client';

export const groupItem: GroupItem = {
  id: 'group-1',
  scheme: 'farm',
  status: 'new',
  risk_level: 'high',
  risk_score: 1,
  risk_parts: {},
  detected_at: '2026-10-07T10:00:00Z',
  first_event_at: '2026-10-07T10:00:00Z',
  last_event_at: '2026-10-07T10:30:00Z',
  features: { accounts_count: 1, total_amount: 100, span_minutes: 30 },
};
