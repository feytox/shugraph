export const panelWidthPx = { initial: 280, min: 220, max: 480, keyboardStep: 16 };

export const clampPanelWidth = (width: number) =>
  Math.min(panelWidthPx.max, Math.max(panelWidthPx.min, width));
