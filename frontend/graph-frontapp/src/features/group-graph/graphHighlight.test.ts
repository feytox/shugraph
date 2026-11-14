import { describe, expect, it } from 'vitest';
import cytoscape from 'cytoscape';
import { mockGroups } from '@/mocks/groupGraphs';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { fadedClass, highlightLinksOnTap } from '@/features/group-graph/graphHighlight';

function farmGraph() {
  const farm = mockGroups.find((group) => group.item.scheme === 'farm')!;
  const cy = cytoscape({ headless: true, elements: toGraphElements('farm', farm.graph) });
  highlightLinksOnTap(cy);
  return cy;
}

const brightIds = (cy: cytoscape.Core) =>
  cy
    .elements()
    .filter((element) => !element.hasClass(fadedClass))
    .map((element) => element.id())
    .sort();

describe('highlightLinksOnTap', () => {
  it('клик по узлу оставляет яркими только его связи и соседей', () => {
    const cy = farmGraph();
    cy.getElementById('account:acc-ravil').emit('tap');
    expect(brightIds(cy)).toEqual(
      [
        'account:acc-collector',
        'account:acc-ravil',
        'device:dev-common',
        'login:evt-df-with-0001:device',
        'transfer:evt-df-with-0004',
      ].sort()
    );
  });

  it('клик по другому узлу переносит подсветку', () => {
    const cy = farmGraph();
    cy.getElementById('account:acc-ravil').emit('tap');
    cy.getElementById('account:acc-dima').emit('tap');
    expect(brightIds(cy)).toContain('account:acc-dima');
    expect(brightIds(cy)).not.toContain('account:acc-ravil');
  });

  it('клик по пустому месту снимает подсветку', () => {
    const cy = farmGraph();
    cy.getElementById('account:acc-ravil').emit('tap');
    cy.emit('tap');
    expect(cy.elements(`.${fadedClass}`)).toHaveLength(0);
  });
});
