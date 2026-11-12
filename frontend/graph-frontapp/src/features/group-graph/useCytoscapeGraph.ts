import { useEffect, useRef, type RefObject } from 'react';
import { useMantineTheme } from '@mantine/core';
import cytoscape, { type ElementsDefinition, type Position } from 'cytoscape';
import type { GroupItem } from '@/api/client';
import { highlightLinksOnTap } from '@/features/group-graph/graphHighlight';
import { graphLayout } from '@/features/group-graph/graphLayout';
import { graphStylesheet } from '@/features/group-graph/graphStyle';
import { keepFittedOnResize } from '@/features/group-graph/keepFittedOnResize';

type Viewport = { zoom: number; pan: Position };

export function useCytoscapeGraph(
  container: RefObject<HTMLDivElement | null>,
  scheme: GroupItem['scheme'],
  elements: ElementsDefinition
) {
  const theme = useMantineTheme();
  const userViewport = useRef<Viewport | null>(null);

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
    if (userViewport.current) cy.viewport(userViewport.current);
    highlightLinksOnTap(cy);
    const fitting = keepFittedOnResize(cy, element, {
      userAlreadyMovedViewport: userViewport.current !== null,
    });

    return () => {
      userViewport.current = fitting.userMovedViewport()
        ? { zoom: cy.zoom(), pan: cy.pan() }
        : null;
      fitting.stop();
      cy.destroy();
    };
  }, [container, scheme, elements, theme]);
}
