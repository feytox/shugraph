import { fireEvent, render, screen } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { beforeEach, describe, expect, it } from 'vitest';
import { SidePanel } from '@/shared/SidePanel';

function renderPanel() {
  return render(
    <MantineProvider>
      <SidePanel side="left" title="Группы" storageKey="test-panel">
        <p>содержимое</p>
      </SidePanel>
    </MantineProvider>
  );
}

const contentIsInert = () => screen.getByText('содержимое').closest('[inert]') !== null;

describe('SidePanel', () => {
  beforeEach(() => localStorage.clear());

  it('сворачивается и разворачивается', () => {
    renderPanel();
    expect(screen.getByRole('heading', { name: 'Группы' })).toBeVisible();
    expect(contentIsInert()).toBe(false);
    fireEvent.click(screen.getByRole('button', { name: 'Скрыть: Группы' }));
    expect(contentIsInert()).toBe(true);
    fireEvent.click(screen.getByRole('button', { name: 'Показать: Группы' }));
    expect(contentIsInert()).toBe(false);
  });

  it('помнит, что панель свёрнута, после перезагрузки', () => {
    const first = renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'Скрыть: Группы' }));
    first.unmount();
    renderPanel();
    expect(screen.getByRole('button', { name: 'Показать: Группы' })).toBeVisible();
    expect(contentIsInert()).toBe(true);
  });
});
