import type { Core } from 'cytoscape';

type RendererWithSyncRender = { renderer?: () => { render?: () => void } };

export const redrawBeforePaint = (cy: Core) =>
  (cy as unknown as RendererWithSyncRender).renderer?.().render?.();
