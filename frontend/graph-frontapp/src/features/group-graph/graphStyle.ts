import type { MantineTheme } from '@mantine/core';
import type { StylesheetJson } from 'cytoscape';
import type { GraphNode } from '@/api/client';

export type ShownNodeType = Exclude<GraphNode['type'], 'client'>;

export const nodeKinds = {
  account: { label: 'счёт', color: 'blue', shape: 'ellipse', legendMarker: 'circle' },
  device: { label: 'устройство', color: 'orange', shape: 'round-diamond', legendMarker: 'diamond' },
  network_address: {
    label: 'сетевой адрес',
    color: 'grape',
    shape: 'round-hexagon',
    legendMarker: 'hexagon',
  },
} as const satisfies Record<ShownNodeType, unknown>;

export const nodeColor = (theme: MantineTheme, type: ShownNodeType) =>
  theme.colors[nodeKinds[type].color][6];

export const graphStylesheet = (theme: MantineTheme): StylesheetJson => [
  {
    selector: 'node',
    style: {
      label: 'data(label)',
      color: theme.black,
      'font-size': 12,
      'text-valign': 'bottom',
      'text-margin-y': 6,
    },
  },
  ...(Object.keys(nodeKinds) as ShownNodeType[]).map((type) => ({
    selector: `node[type = "${type}"]`,
    style: { 'background-color': nodeColor(theme, type), shape: nodeKinds[type].shape },
  })),
  {
    selector: 'edge',
    style: {
      label: 'data(label)',
      color: theme.colors.gray[7],
      'font-size': 10,
      width: 2,
      'curve-style': 'bezier',
      'line-color': theme.colors.gray[6],
      'target-arrow-color': theme.colors.gray[6],
      'target-arrow-shape': 'triangle',
      'text-rotation': 'autorotate',
      'text-background-color': theme.white,
      'text-background-opacity': 1,
    },
  },
  {
    selector: 'edge[type = "login"]',
    style: {
      'line-style': 'dashed',
      'line-color': theme.colors.gray[4],
      'target-arrow-shape': 'none',
    },
  },
];
