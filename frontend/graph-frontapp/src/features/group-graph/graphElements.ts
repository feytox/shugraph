import type { ElementsDefinition } from 'cytoscape';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import type { GroupGraph, GroupItem } from '@/api/client';
import { formatRub } from '@/shared/money';

dayjs.extend(utc);

const time = (occurredAt: string) => dayjs.utc(occurredAt).format('HH:mm');

export function toGraphElements(
  scheme: GroupItem['scheme'],
  graph: GroupGraph
): ElementsDefinition {
  const hiddenTypes = scheme === 'farm' ? ['client', 'network_address'] : ['client'];
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
          label: `${formatRub(transfer.amount)} · ${time(transfer.occurred_at)}`,
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
            label: time(login.occurred_at),
          },
        })),
    ],
  };
}
