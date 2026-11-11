import { describe, expect, it } from 'vitest';
import type { GroupGraph, GroupItem } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';
import { toGraphElements } from '@/features/group-graph/graphElements';

const nbsp = '\u00a0';

const elementsOf = (scheme: GroupItem['scheme']) => {
  const group = mockGroups.find((mock) => mock.item.scheme === scheme)!;
  return toGraphElements(scheme, group.graph);
};

const count = (items: { data: { [key: string]: unknown } }[]) =>
  items.reduce<Record<string, number>>((counts, item) => {
    const type = String(item.data.type);
    counts[type] = (counts[type] ?? 0) + 1;
    return counts;
  }, {});

const label = (elements: ReturnType<typeof toGraphElements>, id: string) =>
  elements.edges.find((edge) => edge.data.id === id)?.data.label;

describe('toGraphElements', () => {
  it('ферма: счета и общее устройство, без клиентов и адресов', () => {
    const elements = elementsOf('farm');
    expect(count(elements.nodes)).toEqual({ account: 4, device: 1 });
    expect(count(elements.edges)).toEqual({ transfer: 3, login: 3 });
    expect(label(elements, 'transfer:evt-df-with-0004')).toBe(`52${nbsp}000${nbsp}₽ · 10:20`);
    expect(label(elements, 'login:evt-df-with-0001:device')).toBe('10:01');
  });

  it('адрес: общий адрес и устройства остаются, клиентов нет', () => {
    const elements = elementsOf('address');
    expect(count(elements.nodes)).toEqual({ account: 6, device: 3, network_address: 1 });
    expect(count(elements.edges)).toEqual({ transfer: 5, login: 10 });
  });

  it('цепочка: только счета и переводы', () => {
    const elements = elementsOf('chain');
    expect(count(elements.nodes)).toEqual({ account: 5 });
    expect(count(elements.edges)).toEqual({ transfer: 4 });
  });

  it('кольцо: повторные переводы между теми же счетами остаются отдельными рёбрами', () => {
    const elements = elementsOf('ring');
    expect(count(elements.nodes)).toEqual({ account: 4 });
    expect(new Set(elements.edges.map((edge) => edge.data.id)).size).toBe(8);
  });

  it('узел подписан исходным id', () => {
    const node = elementsOf('farm').nodes.find((item) => item.data.id === 'device:dev-common');
    expect(node?.data.label).toBe('dev-common');
  });

  it('сумма с копейками показывает копейки, время в UTC', () => {
    const graph: GroupGraph = {
      group_id: 'group-1',
      nodes: [
        { id: 'account:a', type: 'account', entity_id: 'a' },
        { id: 'account:b', type: 'account', entity_id: 'b' },
      ],
      ownerships: [],
      logins: [],
      transfers: [
        {
          id: 'transfer:evt-1',
          event_id: 'evt-1',
          occurred_at: '2026-10-28T23:59:00+03:00',
          source: 'account:a',
          target: 'account:b',
          amount: 123450,
        },
      ],
    };
    expect(label(toGraphElements('chain', graph), 'transfer:evt-1')).toBe(
      `1${nbsp}234,5${nbsp}₽ · 20:59`
    );
  });
});
