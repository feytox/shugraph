import { useEffect, type RefObject } from 'react';
import { useMantineTheme } from '@mantine/core';
import cytoscape from 'cytoscape';
import type { GroupGraph, GroupItem } from '@/api/client';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { highlightLinksOnTap } from '@/features/group-graph/graphHighlight';
import { graphLayout } from '@/features/group-graph/graphLayout';
import { graphStylesheet } from '@/features/group-graph/graphStyle';
import { keepFittedOnResize } from '@/features/group-graph/keepFittedOnResize';

export function useCytoscapeGraph(
  container: RefObject<HTMLDivElement | null>,
  scheme: GroupItem['scheme'],
  graph: GroupGraph
) {
  const theme = useMantineTheme();

  useEffect(() => {
    const element = container.current;
    if (!element) return;

    const cy = cytoscape({
      container: element,
      elements: toGraphElements(scheme, graph),
      style: graphStylesheet(theme),
      autounselectify: true,
      boxSelectionEnabled: false,
    });
    cy.layout(graphLayout(scheme, cy)).run();
    highlightLinksOnTap(cy);
    const stopFitting = keepFittedOnResize(cy, element);

    return () => {
      stopFitting();
      cy.destroy();
    };
  }, [container, scheme, graph, theme]);
}
