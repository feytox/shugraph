import type { EdgeDefinition, ElementsDefinition, Position } from 'cytoscape';
import type { GroupItem } from '@/api/client';

type Positions = Record<string, Position>;
type PointType = 'device' | 'network_address';
type SchemeLayout = (graph: GraphQuery) => Positions;

const distancePx = {
  betweenChainAccounts: 200,
  ringRadius: 200,
  hubToSenders: 260,
  sendersToReceivers: 260,
  betweenAccountRows: 100,
  hubToSecondaryPoints: 150,
  secondaryPointBelowItsAccounts: 45,
};

class GraphQuery {
  private readonly elements: ElementsDefinition;

  constructor(elements: ElementsDefinition) {
    this.elements = elements;
  }

  idsOf(type: 'account' | PointType): string[] {
    return this.elements.nodes
      .filter((node) => node.data.type === type)
      .map((node) => node.data.id!);
  }

  linksInto(id: string): EdgeDefinition[] {
    return this.elements.edges.filter((edge) => edge.data.target === id);
  }

  receivesTransfers(id: string): boolean {
    return this.linksInto(id).some((edge) => edge.data.type === 'transfer');
  }

  busiestPoint(type: PointType): string {
    return this.idsOf(type).reduce((busiest, id) =>
      this.linksInto(id).length > this.linksInto(busiest).length ? id : busiest
    );
  }

  accountsInFirstEventOrder(): string[] {
    const byTime = [...this.elements.edges].sort(
      (left, right) => Date.parse(left.data.occurredAt) - Date.parse(right.data.occurredAt)
    );
    const order = [...new Set(byTime.flatMap((edge) => [edge.data.source, edge.data.target]))];
    const rank = (id: string) => (order.includes(id) ? order.indexOf(id) : order.length);
    return this.idsOf('account').sort((left, right) => rank(left) - rank(right));
  }
}

const line: SchemeLayout = (graph) =>
  Object.fromEntries(
    graph
      .accountsInFirstEventOrder()
      .map((id, index) => [id, { x: index * distancePx.betweenChainAccounts, y: 0 }])
  );

function pointOnCircle(index: number, count: number): Position {
  const angleFromTop = -Math.PI / 2 + (index * 2 * Math.PI) / count;
  return {
    x: distancePx.ringRadius * Math.cos(angleFromTop),
    y: distancePx.ringRadius * Math.sin(angleFromTop),
  };
}

const circle: SchemeLayout = (graph) => {
  const accounts = graph.accountsInFirstEventOrder();
  return Object.fromEntries(
    accounts.map((id, index) => [id, pointOnCircle(index, accounts.length)])
  );
};

const columnAt = (x: number, ids: string[]): Positions =>
  Object.fromEntries(
    ids.map((id, index) => [
      id,
      { x, y: (index - (ids.length - 1) / 2) * distancePx.betweenAccountRows },
    ])
  );

const starAround =
  (hubType: PointType): SchemeLayout =>
  (graph) => {
    const hub = graph.busiestPoint(hubType);
    const accounts = graph.accountsInFirstEventOrder();
    const hubX = -distancePx.hubToSenders;
    const accountPositions = {
      ...columnAt(
        0,
        accounts.filter((id) => !graph.receivesTransfers(id))
      ),
      ...columnAt(
        distancePx.sendersToReceivers,
        accounts.filter((id) => graph.receivesTransfers(id))
      ),
    };

    const besideItsAccounts = (id: string): Position => {
      const rows = graph.linksInto(id).map((edge) => accountPositions[edge.data.source].y);
      const middleRow = rows.reduce((sum, y) => sum + y, 0) / rows.length;
      return {
        x: hubX + distancePx.hubToSecondaryPoints,
        y: middleRow + distancePx.secondaryPointBelowItsAccounts,
      };
    };
    const secondaryPoints = [...graph.idsOf('device'), ...graph.idsOf('network_address')].filter(
      (id) => id !== hub
    );

    return {
      [hub]: { x: hubX, y: 0 },
      ...accountPositions,
      ...Object.fromEntries(secondaryPoints.map((id) => [id, besideItsAccounts(id)])),
    };
  };

const layoutByScheme: Record<GroupItem['scheme'], SchemeLayout> = {
  chain: line,
  ring: circle,
  farm: starAround('device'),
  address: starAround('network_address'),
};

export const layoutGraph = (scheme: GroupItem['scheme'], elements: ElementsDefinition): Positions =>
  layoutByScheme[scheme](new GraphQuery(elements));
