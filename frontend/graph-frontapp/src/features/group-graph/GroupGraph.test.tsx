import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import cytoscape, { type CytoscapeOptions, type ElementsDefinition } from 'cytoscape';
import { http, HttpResponse } from 'msw';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { server } from '@/mocks/server';
import type { GroupItem } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';
import { GroupGraph } from '@/features/group-graph/GroupGraph';
import { graphLayout } from '@/features/group-graph/graphLayout';

const cy = vi.hoisted(() => {
  const layout = { run: vi.fn() };
  return {
    layout: vi.fn(() => layout),
    destroy: vi.fn(),
    on: vi.fn(),
    one: vi.fn(),
    zoom: vi.fn(() => 2),
    pan: vi.fn(() => ({ x: 5, y: 7 })),
    viewport: vi.fn(),
    runLayout: layout.run,
  };
});

vi.mock('cytoscape', () => ({ default: vi.fn(() => cy) }));
vi.mock('@/features/group-graph/graphLayout', () => ({
  graphLayout: vi.fn(() => ({ name: 'grid' })),
  graphPaddingPx: 0,
}));

const farm = mockGroups.find((group) => group.item.scheme === 'farm')!;
const graphUrl = `*/api/v1/groups/${farm.item.id}/graph`;

function graphFor(scheme: GroupItem['scheme'], lastEventAt: string, client: QueryClient) {
  const group = mockGroups.find((mock) => mock.item.scheme === scheme)!;
  return (
    <MantineProvider>
      <QueryClientProvider client={client}>
        <GroupGraph groupId={group.item.id} scheme={scheme} lastEventAt={lastEventAt} />
      </QueryClientProvider>
    </MantineProvider>
  );
}

function renderGraph(scheme: GroupItem['scheme'] = 'farm') {
  return render(graphFor(scheme, farm.item.last_event_at, new QueryClient()));
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

  it('в легенде только то, что есть на графе фермы', async () => {
    renderGraph();
    for (const label of ['счёт', 'устройство', 'перевод', 'вход', 'время в UTC']) {
      expect(await screen.findByText(label)).toBeVisible();
    }
    expect(screen.queryByText('сетевой адрес')).not.toBeInTheDocument();
  });

  it('в легенде цепочки только счета и переводы', async () => {
    renderGraph('chain');
    expect(await screen.findByText('перевод')).toBeVisible();
    for (const label of ['устройство', 'сетевой адрес', 'вход']) {
      expect(screen.queryByText(label)).not.toBeInTheDocument();
    }
  });

  it('перезапрашивает граф, когда в группу пришло новое событие, и не прячет старый', async () => {
    const client = new QueryClient();
    const { rerender } = render(graphFor('farm', farm.item.last_event_at, client));
    await waitFor(() => expect(cytoscape).toHaveBeenCalledTimes(1));
    rerender(graphFor('farm', '2026-11-12T10:45:00Z', client));
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
    await waitFor(() => expect(cytoscape).toHaveBeenCalledTimes(2));
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

  it('при ошибке обновления оставляет последний граф и сообщает об этом', async () => {
    const client = new QueryClient();
    const { rerender } = render(graphFor('farm', farm.item.last_event_at, client));
    await waitFor(() => expect(cytoscape).toHaveBeenCalledTimes(1));
    server.use(http.get(graphUrl, () => new HttpResponse(null, { status: 500 })));
    rerender(graphFor('farm', '2026-11-12T10:45:00Z', client));
    expect(await screen.findByRole('alert')).toHaveTextContent('Не удалось обновить граф');
    expect(
      screen.getByRole('button', { name: 'Повторить: Не удалось обновить граф' })
    ).toBeVisible();
    expect(screen.getByRole('img', { name: 'Граф группы' })).toBeInTheDocument();
    expect(screen.getByText('устройство')).toBeVisible();
    expect(cytoscape).toHaveBeenCalledTimes(1);
  });

  it('после обновления графа той же группы возвращает зум и сдвиг пользователя', async () => {
    cy.one.mockImplementationOnce((_gestures: string, onGesture: () => void) => onGesture());
    const client = new QueryClient();
    const { rerender } = render(graphFor('farm', farm.item.last_event_at, client));
    await waitFor(() => expect(cytoscape).toHaveBeenCalledTimes(1));
    rerender(graphFor('farm', '2026-11-12T10:45:00Z', client));
    await waitFor(() => expect(cytoscape).toHaveBeenCalledTimes(2));
    expect(cy.viewport).toHaveBeenCalledWith({ zoom: 2, pan: { x: 5, y: 7 } });
  });

  it('убирает прежний граф, если при обновлении группа пропала', async () => {
    const client = new QueryClient();
    const { rerender } = render(graphFor('farm', farm.item.last_event_at, client));
    await waitFor(() => expect(cytoscape).toHaveBeenCalledTimes(1));
    server.use(
      http.get(graphUrl, () =>
        HttpResponse.json(
          { type: 'about:blank', title: 'Not Found', status: 404, detail: 'group not found' },
          { status: 404, headers: { 'Content-Type': 'application/problem+json' } }
        )
      )
    );
    rerender(graphFor('farm', '2026-11-12T10:45:00Z', client));
    expect(await screen.findByText('Граф группы не найден')).toBeVisible();
    expect(screen.queryByRole('img', { name: 'Граф группы' })).not.toBeInTheDocument();
  });

  it('оставляет легенду с пометкой UTC, даже когда графа нет', async () => {
    server.use(http.get(graphUrl, () => new HttpResponse(null, { status: 500 })));
    renderGraph();
    expect(await screen.findByText('Не удалось загрузить граф')).toBeVisible();
    expect(screen.getByText('время в UTC')).toBeVisible();
    expect(screen.queryByText('счёт')).not.toBeInTheDocument();
  });

  it('при ошибке сервера даёт повторить загрузку', async () => {
    server.use(http.get(graphUrl, () => new HttpResponse(null, { status: 500 }), { once: true }));
    renderGraph();
    expect(await screen.findByText('Не удалось загрузить граф')).toBeVisible();
    fireEvent.click(screen.getByRole('button', { name: 'Повторить: Не удалось загрузить граф' }));
    await waitFor(() => expect(cytoscape).toHaveBeenCalled());
  });
});
