import type { Core } from 'cytoscape';
import { redrawBeforePaint } from '@/features/group-graph/cytoscapeRedraw';
import { graphPaddingPx } from '@/features/group-graph/graphLayout';

const userViewportGestures = 'dragpan pinchzoom scrollzoom';

export function keepFittedOnResize(cy: Core, container: HTMLElement) {
  let userMovedViewport = false;
  cy.one(userViewportGestures, () => {
    userMovedViewport = true;
  });

  const observer = new ResizeObserver(() => {
    cy.resize();
    if (!userMovedViewport) cy.fit(undefined, graphPaddingPx);
    redrawBeforePaint(cy);
  });
  observer.observe(container);
  return () => observer.disconnect();
}
