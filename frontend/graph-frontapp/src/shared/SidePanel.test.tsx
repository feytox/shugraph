import { fireEvent, render, screen } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { beforeEach, describe, expect, it } from 'vitest';
import { panelWidthPx } from '@/shared/panelWidth';
import { SidePanel } from '@/shared/SidePanel';

function renderPanel(side: 'left' | 'right' = 'left') {
  return render(
    <MantineProvider>
      <SidePanel side={side} title="Группы" storageKey="test-panel">
        <p>содержимое</p>
      </SidePanel>
    </MantineProvider>
  );
}

const separator = () => screen.getByRole('separator', { name: 'Ширина: Группы' });
const width = () => Number(separator().getAttribute('aria-valuenow'));

describe('SidePanel', () => {
  beforeEach(() => localStorage.clear());

  it('сворачивается и разворачивается', () => {
    renderPanel();
    expect(screen.getByRole('heading', { name: 'Группы' })).toBeVisible();
    fireEvent.click(screen.getByRole('button', { name: 'Скрыть: Группы' }));
    expect(screen.queryByText('содержимое')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Показать: Группы' }));
    expect(screen.getByText('содержимое')).toBeVisible();
  });

  it('меняет ширину перетаскиванием края', () => {
    renderPanel();
    fireEvent.pointerDown(separator(), { clientX: 100, pointerId: 1 });
    fireEvent.pointerMove(separator(), { clientX: 160, pointerId: 1 });
    fireEvent.pointerUp(separator(), { pointerId: 1 });
    expect(width()).toBe(panelWidthPx.initial + 60);
    fireEvent.pointerMove(separator(), { clientX: 400, pointerId: 1 });
    expect(width()).toBe(panelWidthPx.initial + 60);
  });

  it('меняет ширину стрелками и не выходит за пределы', () => {
    renderPanel();
    fireEvent.keyDown(separator(), { key: 'ArrowRight' });
    expect(width()).toBe(panelWidthPx.initial + panelWidthPx.keyboardStep);
    for (let press = 0; press < 50; press++) fireEvent.keyDown(separator(), { key: 'ArrowRight' });
    expect(width()).toBe(panelWidthPx.max);
    for (let press = 0; press < 50; press++) fireEvent.keyDown(separator(), { key: 'ArrowLeft' });
    expect(width()).toBe(panelWidthPx.min);
  });

  it('сохранённая ширина меньше минимума поднимается до минимума', () => {
    localStorage.setItem('test-panel-width', String(panelWidthPx.min - 80));
    renderPanel();
    expect(width()).toBe(panelWidthPx.min);
  });

  it('правая панель растёт, когда край тянут влево', () => {
    renderPanel('right');
    fireEvent.keyDown(separator(), { key: 'ArrowLeft' });
    expect(width()).toBe(panelWidthPx.initial + panelWidthPx.keyboardStep);
  });

  it('помнит ширину и свёрнутость после перезагрузки', () => {
    const first = renderPanel();
    fireEvent.keyDown(separator(), { key: 'ArrowRight' });
    first.unmount();
    renderPanel();
    expect(width()).toBe(panelWidthPx.initial + panelWidthPx.keyboardStep);

    fireEvent.click(screen.getByRole('button', { name: 'Скрыть: Группы' }));
    screen.getByRole('button', { name: 'Показать: Группы' });
  });
});
