import type { Core, EventObject } from 'cytoscape';

export const fadedClass = 'faded';

export function highlightLinksOnTap(cy: Core) {
  cy.on('tap', 'node', (event: EventObject) => {
    cy.elements().addClass(fadedClass);
    event.target.closedNeighborhood().removeClass(fadedClass);
  });
  cy.on('tap', (event: EventObject) => {
    if (event.target === cy) cy.elements().removeClass(fadedClass);
  });
}
