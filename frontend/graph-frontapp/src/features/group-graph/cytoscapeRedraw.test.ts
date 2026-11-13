import { describe, expect, it } from 'vitest';
import cytoscape from 'cytoscape';

describe('redrawBeforePaint', () => {
  it('опирается на синхронный render рендерера canvas, который есть в этой версии Cytoscape', () => {
    const canvasRenderer = cytoscape('renderer', 'canvas') as { prototype: { render?: unknown } };
    expect(typeof canvasRenderer.prototype.render).toBe('function');
  });
});
