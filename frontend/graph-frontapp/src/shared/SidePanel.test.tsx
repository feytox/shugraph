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

  it('помнит, что панель свёрнута, после перезагрузки', () => {
    const first = renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'Скрыть: Группы' }));
    first.unmount();
    renderPanel();
    expect(screen.getByRole('button', { name: 'Показать: Группы' })).toBeVisible();
    expect(screen.queryByText('содержимое')).not.toBeInTheDocument();
  });
});
