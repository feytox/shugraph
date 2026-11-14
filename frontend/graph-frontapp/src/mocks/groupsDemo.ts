import type { GroupItem } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';

export function createGroupsDemo() {
  let startedAt: number | undefined;
  function elapsed() {
    startedAt ??= Date.now();
    return Date.now() - startedAt;
  }

  return {
    getItems(): GroupItem[] {
      const step = Math.min(3, Math.floor(elapsed() / 10000));
      const generated: GroupItem[] = Array.from({ length: 25 + step }, (_, index) => {
        const number = 25 + step - index;
        const updated = number === 25;
        return {
          id: `00000000-0000-4000-8000-${String(number).padStart(12, '0')}`,
          scheme: (['farm', 'address', 'chain', 'ring'] as const)[(number - 1) % 4],
          status: 'new',
          risk_level: 'high',
          risk_score: updated ? 87 + step : 60 + number,
          risk_parts: {},
          detected_at: `2026-10-07T10:00:${String(number).padStart(2, '0')}Z`,
          first_event_at: '2026-10-07T10:00:00Z',
          last_event_at: updated
            ? `2026-10-07T10:30:${String(45 + step).padStart(2, '0')}Z`
            : '2026-10-07T10:30:45Z',
          features: {
            accounts_count: number,
            total_amount: updated ? 12345 + step * 10000 : number * 10000,
            span_minutes: 30,
          },
        };
      });
      return [...mockGroups.map((group) => group.item), ...generated];
    },
  };
}
