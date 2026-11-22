import { describe, expect, it } from 'vitest';
import { apiClient } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';

const minutesBetween = (from: string, to: string) => (Date.parse(to) - Date.parse(from)) / 60000;

describe('моки графов групп', () => {
  it('есть группа каждой схемы', () => {
    expect(mockGroups.map((group) => group.item.scheme).sort()).toEqual([
      'address',
      'chain',
      'farm',
      'ring',
    ]);
  });

  describe.each(mockGroups.map((group) => [group.item.scheme, group] as const))(
    '%s',
    (_, { item, graph }) => {
      it('граф принадлежит своей группе', () => {
        expect(graph.group_id).toBe(item.id);
      });

      it('id узла собран из типа и исходного id', () => {
        for (const node of graph.nodes) expect(node.id).toBe(`${node.type}:${node.entity_id}`);
      });

      it('рёбра ведут только в узлы графа', () => {
        const ids = new Set(graph.nodes.map((node) => node.id));
        for (const edge of [...graph.ownerships, ...graph.logins, ...graph.transfers]) {
          expect(ids).toContain(edge.source);
          expect(ids).toContain(edge.target);
        }
      });

      it('числа в списке совпадают с графом', () => {
        const times = [...graph.logins, ...graph.transfers].map((edge) => edge.occurred_at).sort();
        expect(item.features.accounts_count).toBe(
          graph.nodes.filter((node) => node.type === 'account').length
        );
        expect(item.features.total_amount).toBe(
          graph.transfers.reduce((sum, transfer) => sum + transfer.amount, 0)
        );
        expect(item.first_event_at).toBe(times[0]);
        expect(item.last_event_at).toBe(times[times.length - 1]);
        expect(item.features.span_minutes).toBe(minutesBetween(times[0], times[times.length - 1]));
      });
    }
  );

  it('ручка отдаёт граф каждой группы', async () => {
    for (const { item, graph } of mockGroups) {
      const { data, response } = await apiClient.GET('/api/v1/groups/{group_id}/graph', {
        params: { path: { group_id: item.id } },
      });
      expect(response.status).toBe(200);
      expect(data).toEqual(graph);
    }
  });

  it('ручка отвечает 404 на неизвестную группу', async () => {
    const { error, response } = await apiClient.GET('/api/v1/groups/{group_id}/graph', {
      params: { path: { group_id: '00000000-0000-4000-8000-999999999999' } },
    });
    expect(response.status).toBe(404);
    expect(error?.detail).toBe('group not found');
  });
});
