import type { MantineTheme } from '@mantine/core';
import type { Css, StylesheetJson } from 'cytoscape';
import { fadedClass } from '@/features/group-graph/graphHighlight';
import { nodeKinds, shownNodeTypes, type ShownNodeType } from '@/features/group-graph/nodeKinds';

const graphSizesPx = {
  nodeHeight: 30,
  nodeLabelFont: 12,
  nodeLabelGap: 6,
  edgeLabelFont: 10,
  edgeWidth: 2,
};
const fadedNodeOpacity = 0.15;

const nodeColor = (theme: MantineTheme, type: ShownNodeType) =>
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
      'font-size': graphSizesPx.nodeLabelFont,
      'text-valign': 'bottom',
      'text-margin-y': graphSizesPx.nodeLabelGap,
    },
  },
  ...shownNodeTypes.map((type) => ({
    selector: `node[type = "${type}"]`,
    style: {
      'background-color': nodeColor(theme, type),
      shape: nodeKinds[type].shape,
      width: graphSizesPx.nodeHeight * nodeKinds[type].widthToHeight,
      height: graphSizesPx.nodeHeight,
    },
  })),
  {
    selector: 'edge',
    style: {
      label: 'data(label)',
      color: theme.colors.gray[7],
      'font-size': graphSizesPx.edgeLabelFont,
      width: graphSizesPx.edgeWidth,
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
  { selector: `node.${fadedClass}`, style: { opacity: fadedNodeOpacity } },
  {
    selector: `edge.${fadedClass}`,
    style: {
      color: theme.colors.gray[3],
      'line-color': theme.colors.gray[2],
      'target-arrow-color': theme.colors.gray[2],
    },
  },
];
