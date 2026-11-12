import { afterEach, describe, expect, it, vi } from 'vitest';
import cytoscape from 'cytoscape';
import { keepFittedOnResize } from '@/features/group-graph/keepFittedOnResize';

let notifyResize = () => {};

vi.stubGlobal(
  'ResizeObserver',
  class {
    constructor(callback: () => void) {
      notifyResize = callback;
    }
    observe() {}
    disconnect() {}
  }
);

function graphWithFitSpy() {
  const cy = cytoscape({ headless: true, elements: [{ data: { id: 'a' } }] });
  keepFittedOnResize(cy, document.createElement('div'));
  return { cy, fit: vi.spyOn(cy, 'fit') };
}

describe('keepFittedOnResize', () => {
  afterEach(() => vi.clearAllMocks());

  it('подгоняет граф под новый размер, пока пользователь его не двигал', () => {
    const { fit } = graphWithFitSpy();
    notifyResize();
    expect(fit).toHaveBeenCalledTimes(1);
  });

  it('не сбрасывает зум и сдвиг после жеста пользователя', () => {
    const { cy, fit } = graphWithFitSpy();
    cy.emit('scrollzoom');
    notifyResize();
    expect(fit).not.toHaveBeenCalled();
  });

  it('не подгоняет обновлённый граф, если пользователь двигал прежний', () => {
    const cy = cytoscape({ headless: true, elements: [{ data: { id: 'a' } }] });
    const fitting = keepFittedOnResize(cy, document.createElement('div'), {
      userAlreadyMovedViewport: true,
    });
    const fit = vi.spyOn(cy, 'fit');
    notifyResize();
    expect(fit).not.toHaveBeenCalled();
    expect(fitting.userMovedViewport()).toBe(true);
  });
});
