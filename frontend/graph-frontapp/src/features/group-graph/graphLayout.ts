import type {
  BreadthFirstLayoutOptions,
  CircleLayoutOptions,
  Core,
  NodeCollection,
  NodeSingular,
  SortingFunction,
} from 'cytoscape';
import type { GroupItem } from '@/api/client';

type PointType = 'device' | 'network_address';
type GraphLayoutOptions = (BreadthFirstLayoutOptions | CircleLayoutOptions) & {
  eles?: NodeCollection;
};

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

const bottomUp =
  (sort: SortingFunction): SortingFunction =>
  (left, right) =>
    sort(right, left);

const busiestPoint = (cy: Core, type: PointType) =>
  cy
    .nodes(`[type = "${type}"]`)
    .max((node) => node.degree(false))
    .ele.id();

const shared = { animate: false, nodeDimensionsIncludeLabels: true } as const;

const columnsFrom = (cy: Core, roots: string[], directed: boolean): GraphLayoutOptions => ({
  name: 'breadthfirst',
  roots,
  directed,
  direction: 'rightward',
  depthSort: bottomUp(byFirstEvent(cy)),
  ...shared,
});

const layoutByScheme: Record<GroupItem['scheme'], (cy: Core) => GraphLayoutOptions> = {
  chain: (cy) =>
    columnsFrom(
      cy,
      cy
        .nodes()
        .roots()
        .map((node) => node.id()),
      true
    ),
  ring: (cy) => ({
    name: 'circle',
    eles: cy.nodes().sort(byFirstEvent(cy)),
    startAngle: (3 / 2) * Math.PI,
    clockwise: true,
    ...shared,
  }),
  farm: (cy) => columnsFrom(cy, [busiestPoint(cy, 'device')], false),
  address: (cy) => columnsFrom(cy, [busiestPoint(cy, 'network_address')], false),
};

export const graphLayout = (scheme: GroupItem['scheme'], cy: Core): GraphLayoutOptions =>
  layoutByScheme[scheme](cy);
