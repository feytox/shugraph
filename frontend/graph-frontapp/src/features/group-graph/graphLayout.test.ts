import { describe, expect, it } from 'vitest';
import type { ElementsDefinition } from 'cytoscape';
import type { GroupItem } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { layoutGraph } from '@/features/group-graph/graphLayout';

const elementsOf = (scheme: GroupItem['scheme']) => {
  const group = mockGroups.find((mock) => mock.item.scheme === scheme)!;
  return toGraphElements(scheme, group.graph);
};

const reversed = ({ nodes, edges }: ElementsDefinition): ElementsDefinition => ({
  nodes: [...nodes].reverse(),
  edges: [...edges].reverse(),
});

describe('layoutGraph', () => {
  it.each(['farm', 'address', 'chain', 'ring'] as const)(
    '%s: у каждого узла есть позиция, порядок в ответе не важен',
    (scheme) => {
      const elements = elementsOf(scheme);
      const positions = layoutGraph(scheme, elements);
      expect(Object.keys(positions).sort()).toEqual(
        elements.nodes.map((node) => node.data.id).sort()
      );
      expect(layoutGraph(scheme, reversed(elements))).toEqual(positions);
    }
  );

  it('цепочка: счета в линию по времени переводов', () => {
    const positions = layoutGraph('chain', reversed(elementsOf('chain')));
    expect(['c-1', 'c-2', 'c-3', 'c-4', 'c-5'].map((id) => positions[`account:${id}`])).toEqual([
      { x: 0, y: 0 },
      { x: 200, y: 0 },
      { x: 400, y: 0 },
      { x: 600, y: 0 },
      { x: 800, y: 0 },
    ]);
  });

  it('кольцо: счета по кругу, первый сверху, дальше по часовой стрелке', () => {
    const positions = layoutGraph('ring', elementsOf('ring'));
    const expected = [
      [0, -200],
      [200, 0],
      [0, 200],
      [-200, 0],
    ];
    ['r-1', 'r-2', 'r-3', 'r-4'].forEach((id, index) => {
      expect(positions[`account:${id}`].x).toBeCloseTo(expected[index][0]);
      expect(positions[`account:${id}`].y).toBeCloseTo(expected[index][1]);
    });
  });

  it('ферма: устройство слева, дропы в колонку, сборщик справа', () => {
    const positions = layoutGraph('farm', elementsOf('farm'));
    expect(positions).toEqual({
      'device:dev-common': { x: -260, y: 0 },
      'account:acc-ravil': { x: 0, y: -100 },
      'account:acc-sasha': { x: 0, y: 0 },
      'account:acc-dima': { x: 0, y: 100 },
      'account:acc-collector': { x: 260, y: 0 },
    });
  });

  it('адрес: адрес слева, устройства между ним и своими счетами', () => {
    const positions = layoutGraph('address', elementsOf('address'));
    expect(positions['network_address:178.4.2.10']).toEqual({ x: -260, y: 0 });
    expect(['a-1', 'a-2', 'a-3', 'a-4', 'a-5'].map((id) => positions[`account:${id}`])).toEqual([
      { x: 0, y: -200 },
      { x: 0, y: -100 },
      { x: 0, y: 0 },
      { x: 0, y: 100 },
      { x: 0, y: 200 },
    ]);
    expect(positions['account:a-6']).toEqual({ x: 260, y: 0 });
    expect(positions['device:pc-01']).toEqual({ x: -110, y: -105 });
    expect(positions['device:pc-02']).toEqual({ x: -110, y: 45 });
    expect(positions['device:phone-9c']).toEqual({ x: -110, y: 195 });
  });
});
