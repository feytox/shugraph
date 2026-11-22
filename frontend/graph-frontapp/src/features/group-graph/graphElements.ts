import type { ElementsDefinition } from 'cytoscape';
import type { GraphNode, GroupGraph, GroupItem } from '@/api/client';
import { formatTimeUtc } from '@/shared/dates';
import { formatRub } from '@/shared/money';

const hiddenNodeTypesByScheme: Record<GroupItem['scheme'], GraphNode['type'][]> = {
  farm: ['client', 'network_address'],
  address: ['client'],
  chain: ['client'],
  ring: ['client'],
};

export function toGraphElements(
  scheme: GroupItem['scheme'],
  graph: GroupGraph
): ElementsDefinition {
  const hiddenTypes = hiddenNodeTypesByScheme[scheme];
  const nodes = graph.nodes.filter((node) => !hiddenTypes.includes(node.type));
  const shown = new Set(nodes.map((node) => node.id));

  return {
    nodes: nodes.map((node) => ({
      data: { id: node.id, type: node.type, label: node.entity_id },
    })),
    edges: [
      ...graph.transfers.map((transfer) => ({
        data: {
          id: transfer.id,
          type: 'transfer',
          source: transfer.source,
          target: transfer.target,
          occurredAt: transfer.occurred_at,
          label: `${formatRub(transfer.amount)} · ${formatTimeUtc(transfer.occurred_at)}`,
        },
      })),
      ...graph.logins
        .filter((login) => shown.has(login.target))
        .map((login) => ({
          data: {
            id: login.id,
            type: 'login',
            source: login.source,
            target: login.target,
            occurredAt: login.occurred_at,
            label: formatTimeUtc(login.occurred_at),
          },
        })),
    ],
  };
}
