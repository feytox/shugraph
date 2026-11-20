import { describe, expect, it } from 'vitest';
import cytoscape from 'cytoscape';
import { DEFAULT_THEME } from '@mantine/core';
import { mockGroups } from '@/mocks/groupGraphs';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { graphStylesheet } from '@/features/group-graph/graphStyle';

describe('graphStylesheet', () => {
  it('подписи узлов и рёбер рисуются шрифтом темы', () => {
    const theme = { ...DEFAULT_THEME, fontFamily: 'Theme Sans, sans-serif' };
    const farm = mockGroups.find((group) => group.item.scheme === 'farm')!;
    const cy = cytoscape({
      headless: true,
      styleEnabled: true,
      style: graphStylesheet(theme),
      elements: toGraphElements('farm', farm.graph),
    });

    const fonts = new Set(cy.elements().map((element) => element.style('font-family')));
    expect(fonts).toEqual(new Set(['Theme Sans, sans-serif']));
  });
});
