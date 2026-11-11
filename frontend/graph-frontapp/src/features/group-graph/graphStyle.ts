import type { MantineTheme } from '@mantine/core';
import type { Css, StylesheetJson } from 'cytoscape';
import type { GraphNode } from '@/api/client';

export type ShownNodeType = Exclude<GraphNode['type'], 'client'>;

const nodeHeightPx = 30;
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

export const nodeColor = (theme: MantineTheme, type: ShownNodeType) =>
  theme.colors[nodeKinds[type].color][6];

const noTapMarkOnBackground = { 'active-bg-opacity': 0 } as Css.Core;

export const graphStylesheet = (theme: MantineTheme): StylesheetJson => [
  { selector: 'core', style: noTapMarkOnBackground },
  { selector: 'node:active', style: { 'overlay-opacity': 0 } },
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
    style: {
      'background-color': nodeColor(theme, type),
      shape: nodeKinds[type].shape,
      width: nodeHeightPx * nodeKinds[type].widthToHeight,
      height: nodeHeightPx,
    },
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
      events: 'no',
      'text-background-color': theme.colors.gray[0],
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
