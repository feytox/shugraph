import type {
  BreadthFirstLayoutOptions,
  CircleLayoutOptions,
  Core,
  GridLayoutOptions,
  NodeCollection,
  NodeSingular,
  SortingFunction,
} from 'cytoscape';
import type { GroupItem } from '@/api/client';

type PointType = 'device' | 'network_address';
type Cell = { row: number; col: number };
type GraphLayoutOptions = (BreadthFirstLayoutOptions | CircleLayoutOptions | GridLayoutOptions) & {
  eles?: NodeCollection;
};

const starColumn = { hub: 0, participants: 1, receivers: 2 } as const;

function byFirstEvent(cy: Core): SortingFunction {
  const rank = new Map<string, number>();
  cy.edges()
    .sort(
      (left, right) => Date.parse(left.data('occurredAt')) - Date.parse(right.data('occurredAt'))
    )
    .forEach((edge) => {
      for (const id of [edge.source().id(), edge.target().id()]) {
        if (!rank.has(id)) rank.set(id, rank.size);
      }
    });
  const rankOf = (node: NodeSingular) => rank.get(node.id()) ?? rank.size;
  return (left, right) => rankOf(left) - rankOf(right);
}

const busiestPoint = (cy: Core, type: PointType) =>
  cy
    .nodes(`[type = "${type}"]`)
    .max((node) => node.degree(false))
    .ele.id();

export const graphPaddingPx = 30;

const shared = {
  animate: false,
  nodeDimensionsIncludeLabels: true,
  padding: graphPaddingPx,
} as const;

const line = (cy: Core): GraphLayoutOptions => ({
  name: 'breadthfirst',
  roots: cy
    .nodes()
    .roots()
    .map((node) => node.id()),
  directed: true,
  direction: 'rightward',
  ...shared,
});

const circle = (cy: Core): GraphLayoutOptions => ({
  name: 'circle',
  eles: cy.nodes().sort(byFirstEvent(cy)),
  startAngle: (3 / 2) * Math.PI,
  clockwise: true,
  ...shared,
});

const starAround =
  (hubType: PointType) =>
  (cy: Core): GraphLayoutOptions => {
    const inEventOrder = byFirstEvent(cy);
    const hub = cy.getElementById(busiestPoint(cy, hubType));
    const receivers = cy
      .nodes('[type = "account"]')
      .filter((node) => node.incomers('edge[type = "transfer"]').nonempty())
      .sort(inEventOrder);
    const participants = cy.nodes().difference(receivers).difference(hub).sort(inEventOrder);
    const rows = Math.max(participants.length, receivers.length, 1);
    const centeredRow = (index: number, count: number) => Math.floor((rows - count) / 2) + index;

    const cells = new Map<string, Cell>([
      [hub.id(), { row: centeredRow(0, 1), col: starColumn.hub }],
    ]);
    participants.forEach((node, index) => {
      cells.set(node.id(), { row: index, col: starColumn.participants });
    });
    receivers.forEach((node, index) => {
      cells.set(node.id(), {
        row: centeredRow(index, receivers.length),
        col: starColumn.receivers,
      });
    });

    return {
      name: 'grid',
      rows,
      cols: Object.keys(starColumn).length,
      position: (node) => cells.get(node.id())!,
      ...shared,
    };
  };

const layoutByScheme: Record<GroupItem['scheme'], (cy: Core) => GraphLayoutOptions> = {
  chain: line,
  ring: circle,
  farm: starAround('device'),
  address: starAround('network_address'),
};

export const graphLayout = (scheme: GroupItem['scheme'], cy: Core): GraphLayoutOptions =>
  layoutByScheme[scheme](cy);
