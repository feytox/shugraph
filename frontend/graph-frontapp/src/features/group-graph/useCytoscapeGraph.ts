import { useEffect, type RefObject } from 'react';
import { useMantineTheme } from '@mantine/core';
import cytoscape, { type ElementsDefinition } from 'cytoscape';
import type { GroupItem } from '@/api/client';
import { highlightLinksOnTap } from '@/features/group-graph/graphHighlight';
import { graphLayout } from '@/features/group-graph/graphLayout';
import { graphStylesheet } from '@/features/group-graph/graphStyle';
import { keepFittedOnResize } from '@/features/group-graph/keepFittedOnResize';

export function useCytoscapeGraph(
  container: RefObject<HTMLDivElement | null>,
  scheme: GroupItem['scheme'],
  elements: ElementsDefinition
) {
  const theme = useMantineTheme();

  useEffect(() => {
    const element = container.current;
    if (!element) return;

    const cy = cytoscape({
      container: element,
      elements,
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
  }, [container, scheme, elements, theme]);
}
