import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import cytoscape, { type CytoscapeOptions, type ElementsDefinition } from 'cytoscape';
import { http, HttpResponse } from 'msw';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { server } from '@/mocks/server';
import { mockGroups } from '@/mocks/groupGraphs';
import { GroupGraph } from '@/features/group-graph/GroupGraph';
import { graphLayout } from '@/features/group-graph/graphLayout';

const cy = vi.hoisted(() => {
  const layout = { run: vi.fn() };
  return { layout: vi.fn(() => layout), destroy: vi.fn(), on: vi.fn(), runLayout: layout.run };
});

vi.mock('cytoscape', () => ({ default: vi.fn(() => cy) }));
vi.mock('@/features/group-graph/graphLayout', () => ({
  graphLayout: vi.fn(() => ({ name: 'grid' })),
  graphPaddingPx: 0,
}));

const farm = mockGroups.find((group) => group.item.scheme === 'farm')!;
const graphUrl = `*/api/v1/groups/${farm.item.id}/graph`;

function renderGraph() {
  return render(
    <MantineProvider>
      <QueryClientProvider client={new QueryClient()}>
        <GroupGraph groupId={farm.item.id} scheme="farm" />
      </QueryClientProvider>
    </MantineProvider>
  );
}

const passedOptions = () => vi.mocked(cytoscape).mock.calls.at(-1)![0] as CytoscapeOptions;

describe('GroupGraph', () => {
  beforeEach(() => vi.clearAllMocks());

  it('пока граф грузится, показывает загрузку', () => {
    renderGraph();
    expect(screen.getByRole('status')).toHaveTextContent('Загрузка графа');
  });

  it('рисует граф запрошенной группы и раскладывает его по схеме', async () => {
    renderGraph();
    await waitFor(() => expect(cytoscape).toHaveBeenCalled());
    const { elements, container } = passedOptions();
    expect((elements as ElementsDefinition).nodes.map((node) => node.data.id)).toContain(
      'device:dev-common'
    );
    expect(container).toBe(screen.getByRole('img', { name: 'Граф группы' }));
    expect(graphLayout).toHaveBeenCalledWith('farm', cy);
    expect(cy.runLayout).toHaveBeenCalled();
    expect(cy.on).toHaveBeenCalledWith('tap', 'node', expect.any(Function));
  });

  it('показывает легенду', async () => {
    renderGraph();
    for (const label of ['счёт', 'устройство', 'сетевой адрес', 'перевод', 'вход', 'время в UTC']) {
      expect(await screen.findByText(label)).toBeVisible();
    }
  });

  it('убирает граф, когда компонент закрывают', async () => {
    const { unmount } = renderGraph();
    await waitFor(() => expect(cytoscape).toHaveBeenCalled());
    unmount();
    expect(cy.destroy).toHaveBeenCalled();
  });

  it('сообщает, что граф не найден, если группы нет', async () => {
    server.use(http.get(graphUrl, () => new HttpResponse(null, { status: 404 })));
    renderGraph();
    expect(await screen.findByText('Граф группы не найден')).toBeVisible();
    expect(cytoscape).not.toHaveBeenCalled();
  });

  it('при ошибке сервера даёт повторить загрузку', async () => {
    server.use(http.get(graphUrl, () => new HttpResponse(null, { status: 500 }), { once: true }));
    renderGraph();
    expect(await screen.findByText('Не удалось загрузить граф группы')).toBeVisible();
    fireEvent.click(screen.getByRole('button', { name: 'Повторить' }));
    await waitFor(() => expect(cytoscape).toHaveBeenCalled());
  });
});
