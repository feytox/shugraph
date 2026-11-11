import type { GraphNode } from '@/api/client';

export type ShownNodeType = Exclude<GraphNode['type'], 'client'>;

const regularHexagonWidthToHeight = 2 / Math.sqrt(3);

export const nodeKinds = {
  account: {
    label: 'счёт',
    color: 'blue',
    shape: 'ellipse',
    widthToHeight: 1,
    legendMarker: 'circle',
  },
  device: {
    label: 'устройство',
    color: 'orange',
    shape: 'round-diamond',
    widthToHeight: 1,
    legendMarker: 'diamond',
  },
  network_address: {
    label: 'сетевой адрес',
    color: 'grape',
    shape: 'round-hexagon',
    widthToHeight: regularHexagonWidthToHeight,
    legendMarker: 'hexagon',
  },
} as const satisfies Record<ShownNodeType, unknown>;

export const shownNodeTypes = Object.keys(nodeKinds) as ShownNodeType[];
