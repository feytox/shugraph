import { describe, expect, it } from 'vitest';
import cytoscape, { type ElementsDefinition } from 'cytoscape';
import type { GroupItem } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { graphLayout } from '@/features/group-graph/graphLayout';

const elementsOf = (scheme: GroupItem['scheme']) => {
  const group = mockGroups.find((mock) => mock.item.scheme === scheme)!;
  return toGraphElements(scheme, group.graph);
};

const reversed = ({ nodes, edges }: ElementsDefinition): ElementsDefinition => ({
  nodes: [...nodes].reverse(),
  edges: [...edges].reverse(),
});

const containerArea = { x1: 0, y1: 0, w: 1000, h: 600 };

function laidOut(scheme: GroupItem['scheme'], elements = elementsOf(scheme)) {
  const cy = cytoscape({ headless: true, elements });
  cy.layout({ ...graphLayout(scheme, cy), boundingBox: containerArea }).run();
  return (id: string) => cy.getElementById(id).position();
}

const xs = (at: ReturnType<typeof laidOut>, ids: string[]) => ids.map((id) => at(id).x);
const ys = (at: ReturnType<typeof laidOut>, ids: string[]) => ids.map((id) => at(id).y);
const increasing = (values: number[]) =>
  values.every((value, index) => index === 0 || value > values[index - 1]);
const same = (values: number[]) => values.every((value) => Math.abs(value - values[0]) < 0.001);

describe('graphLayout', () => {
  it.each(['farm', 'address', 'chain', 'ring'] as const)(
    '%s: порядок узлов и рёбер в ответе не меняет картинку',
    (scheme) => {
      const elements = elementsOf(scheme);
      const direct = laidOut(scheme, elements);
      const backwards = laidOut(scheme, reversed(elements));
      for (const node of elements.nodes) {
        expect(backwards(node.data.id!)).toEqual(direct(node.data.id!));
      }
    }
  );

  it('цепочка: счета в линию слева направо по ходу денег', () => {
    const at = laidOut('chain', reversed(elementsOf('chain')));
    const chain = ['c-1', 'c-2', 'c-3', 'c-4', 'c-5'].map((id) => `account:${id}`);
    expect(increasing(xs(at, chain))).toBe(true);
    expect(same(ys(at, chain))).toBe(true);
  });

  it('кольцо: первый счёт сверху, дальше по часовой стрелке', () => {
    const at = laidOut('ring');
    const [first, second, third, fourth] = ['r-1', 'r-2', 'r-3', 'r-4'].map((id) =>
      at(`account:${id}`)
    );
    expect(first.y).toBeLessThan(third.y);
    expect(second.x).toBeGreaterThan(fourth.x);
    expect(first.x).toBeCloseTo(third.x);
    expect(second.y).toBeCloseTo(fourth.y);
  });

  it('ферма: устройство слева, дропы колонкой по времени входа, сборщик справа', () => {
    const at = laidOut('farm');
    const drops = ['acc-ravil', 'acc-sasha', 'acc-dima'].map((id) => `account:${id}`);
    expect(same(xs(at, drops))).toBe(true);
    expect(increasing(ys(at, drops))).toBe(true);
    expect(at('device:dev-common').x).toBeLessThan(at(drops[0]).x);
    expect(at(drops[0]).x).toBeLessThan(at('account:acc-collector').x);
  });

  it('адрес: адрес слева, счета колонкой, получатель и устройства справа', () => {
    const at = laidOut('address');
    const senders = ['a-1', 'a-2', 'a-3', 'a-4', 'a-5'].map((id) => `account:${id}`);
    expect(same(xs(at, senders))).toBe(true);
    expect(increasing(ys(at, senders))).toBe(true);
    expect(at('network_address:178.4.2.10').x).toBeLessThan(at(senders[0]).x);
    for (const right of ['account:a-6', 'device:pc-01', 'device:pc-02', 'device:phone-9c']) {
      expect(at(right).x).toBeGreaterThan(at(senders[0]).x);
    }
  });
});
