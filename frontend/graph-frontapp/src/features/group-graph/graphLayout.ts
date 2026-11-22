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
type SchemeLayout = (cy: Core) => GraphLayoutOptions;
type StarNodes = { hub: NodeSingular; participants: NodeCollection; receivers: NodeCollection };

export const graphPaddingPx = 30;
const topOfCircleRad = (3 / 2) * Math.PI;
const starColumn = { hub: 0, participants: 1, receivers: 2 } as const;

const commonLayoutOptions = {
  animate: false,
  nodeDimensionsIncludeLabels: true,
  padding: graphPaddingPx,
} as const;

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

const busiestPoint = (cy: Core, type: PointType): NodeSingular | undefined =>
  cy.nodes(`[type = "${type}"]`).max((node) => node.degree(false)).ele as NodeSingular | undefined;

const lineLayout: SchemeLayout = (cy) => {
  const roots = cy
    .nodes()
    .roots()
    .map((node) => node.id());
  return {
    name: 'breadthfirst',
    roots: roots.length ? roots : undefined,
    directed: true,
    direction: 'rightward',
    ...commonLayoutOptions,
  };
};

const circleLayout: SchemeLayout = (cy) => ({
  name: 'circle',
  eles: cy.nodes().sort(byFirstEvent(cy)),
  startAngle: topOfCircleRad,
  clockwise: true,
  ...commonLayoutOptions,
});

function classifyStarNodes(cy: Core, hub: NodeSingular): StarNodes {
  const inEventOrder = byFirstEvent(cy);
  const receivers = cy
    .nodes('[type = "account"]')
    .filter((node) => node.incomers('edge[type = "transfer"]').nonempty())
    .sort(inEventOrder);
  const participants = cy.nodes().difference(receivers).difference(hub).sort(inEventOrder);
  return { hub, participants, receivers };
}

function starGrid({ hub, participants, receivers }: StarNodes): GraphLayoutOptions {
  const rows = Math.max(participants.length, receivers.length, 1);
  const centeredRow = (index: number, count: number) => Math.floor((rows - count) / 2) + index;
  const cells = new Map<string, Cell>([
    [hub.id(), { row: centeredRow(0, 1), col: starColumn.hub }],
    ...participants.map((node, index): [string, Cell] => [
      node.id(),
      { row: index, col: starColumn.participants },
    ]),
    ...receivers.map((node, index): [string, Cell] => [
      node.id(),
      { row: centeredRow(index, receivers.length), col: starColumn.receivers },
    ]),
  ]);

  return {
    name: 'grid',
    rows,
    cols: Object.keys(starColumn).length,
    position: (node) => cells.get(node.id())!,
    ...commonLayoutOptions,
  };
}

const starLayoutAround =
  (hubType: PointType): SchemeLayout =>
  (cy) => {
    const hub = busiestPoint(cy, hubType);
    return hub ? starGrid(classifyStarNodes(cy, hub)) : lineLayout(cy);
  };

const layoutByScheme: Record<GroupItem['scheme'], SchemeLayout> = {
  chain: lineLayout,
  ring: circleLayout,
  farm: starLayoutAround('device'),
  address: starLayoutAround('network_address'),
};

export const graphLayout = (scheme: GroupItem['scheme'], cy: Core): GraphLayoutOptions =>
  layoutByScheme[scheme](cy);
