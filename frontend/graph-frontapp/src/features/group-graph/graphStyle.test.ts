import { describe, expect, it } from 'vitest';
import cytoscape from 'cytoscape';
import { DEFAULT_THEME, mergeMantineTheme } from '@mantine/core';
import { mockGroups } from '@/mocks/groupGraphs';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { graphStylesheet } from '@/features/group-graph/graphStyle';
import { theme } from '@/theme';

describe('graphStylesheet', () => {
  it('подписи узлов и рёбер рисуются шрифтом темы приложения', () => {
    const farm = mockGroups.find((group) => group.item.scheme === 'farm')!;
    const cy = cytoscape({
      headless: true,
      styleEnabled: true,
      style: graphStylesheet(mergeMantineTheme(DEFAULT_THEME, theme)),
      elements: toGraphElements('farm', farm.graph),
    });

    const fonts = new Set(cy.elements().map((element) => element.style('font-family')));
    expect([...fonts]).toEqual([expect.stringMatching(/^"Golos Text Variable", /)]);
  });
});
