import type { EdgeDefinition, ElementsDefinition, Position } from 'cytoscape';
import type { GroupItem } from '@/api/client';

const chainStep = 200;
const ringRadius = 200;
const hubX = -260;
const sinkX = 260;
const rowStep = 100;
const secondaryPointX = -110;
const secondaryPointDrop = 45;

const byTime = (left: EdgeDefinition, right: EdgeDefinition) =>
  Date.parse(left.data.occurredAt) - Date.parse(right.data.occurredAt);

function inFirstEventOrder(ids: string[], edges: EdgeDefinition[]) {
  const seen = new Set<string>();
  for (const edge of [...edges].sort(byTime)) {
    seen.add(edge.data.source);
    seen.add(edge.data.target);
  }
  const order = [...seen];
  const rank = (id: string) => (order.includes(id) ? order.indexOf(id) : order.length);
  return [...ids].sort((left, right) => rank(left) - rank(right));
}

const column = (ids: string[], x: number, positions: Record<string, Position>) =>
  ids.forEach((id, index) => {
    positions[id] = { x, y: (index - (ids.length - 1) / 2) * rowStep };
  });

function star(
  hubType: string,
  nodeIds: (type: string) => string[],
  edges: EdgeDefinition[]
): Record<string, Position> {
  const positions: Record<string, Position> = {};
  const linksTo = (id: string) => edges.filter((edge) => edge.data.target === id);
  const hub = nodeIds(hubType).reduce((best, id) =>
    linksTo(id).length > linksTo(best).length ? id : best
  );
  const accounts = inFirstEventOrder(nodeIds('account'), edges);
  const receivesTransfers = (id: string) =>
    linksTo(id).some((edge) => edge.data.type === 'transfer');

  positions[hub] = { x: hubX, y: 0 };
  column(
    accounts.filter((id) => !receivesTransfers(id)),
    0,
    positions
  );
  column(accounts.filter(receivesTransfers), sinkX, positions);

  for (const id of [...nodeIds('device'), ...nodeIds('network_address')]) {
    if (id === hub) continue;
    const accountRows = linksTo(id).map((edge) => positions[edge.data.source].y);
    const middle = accountRows.reduce((sum, y) => sum + y, 0) / accountRows.length;
    positions[id] = { x: secondaryPointX, y: middle + secondaryPointDrop };
  }
  return positions;
}

export function layoutGraph(
  scheme: GroupItem['scheme'],
  { nodes, edges }: ElementsDefinition
): Record<string, Position> {
  const nodeIds = (type: string) =>
    nodes.filter((node) => node.data.type === type).map((node) => node.data.id!);
  const accounts = inFirstEventOrder(nodeIds('account'), edges);

  switch (scheme) {
    case 'chain':
      return Object.fromEntries(accounts.map((id, index) => [id, { x: index * chainStep, y: 0 }]));
    case 'ring':
      return Object.fromEntries(
        accounts.map((id, index) => {
          const angle = -Math.PI / 2 + (index * 2 * Math.PI) / accounts.length;
          return [id, { x: ringRadius * Math.cos(angle), y: ringRadius * Math.sin(angle) }];
        })
      );
    case 'farm':
      return star('device', nodeIds, edges);
    case 'address':
      return star('network_address', nodeIds, edges);
  }
}
