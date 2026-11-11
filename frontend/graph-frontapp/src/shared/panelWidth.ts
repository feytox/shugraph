export const panelWidthPx = { initial: 320, min: 300, max: 480, keyboardStep: 16 };

export const clampPanelWidth = (width: number) =>
  Math.min(panelWidthPx.max, Math.max(panelWidthPx.min, width));
