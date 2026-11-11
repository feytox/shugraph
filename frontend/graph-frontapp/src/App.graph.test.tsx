import { fireEvent, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import type { GroupItem, GroupListResponse } from '@/api/client';
import { mockGroups } from '@/mocks/groupGraphs';
import { server } from '@/mocks/server';
import { advance, renderScreen, waitForScreen } from '@/test/renderScreen';

vi.mock('cytoscape', () => ({
  default: vi.fn(() => ({ layout: () => ({ run: vi.fn() }), destroy: vi.fn() })),
}));
vi.mock('@/features/group-graph/graphLayout', () => ({
  graphLayout: vi.fn(() => ({ name: 'grid' })),
}));

const [farm, address, chain, ring] = mockGroups.map((group) => group.item);
let requestedGraphs: string[] = [];

function respond(items: GroupItem[]) {
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json<GroupListResponse>({ items, next_cursor: null })
    )
  );
}

const groupLink = (scheme: string) => screen.getByRole('link', { name: new RegExp(scheme) });
const currentLink = () => screen.getByRole('link', { current: true });
const lastRequestedGraph = () => requestedGraphs.at(-1);

beforeEach(() => {
  requestedGraphs = [];
  server.events.on('request:start', ({ request }) => {
    const match = new URL(request.url).pathname.match(/\/groups\/([^/]+)\/graph$/);
    if (match) requestedGraphs.push(match[1]);
  });
  respond(mockGroups.map((group) => group.item));
});

afterEach(() => {
  server.events.removeAllListeners();
  vi.useRealTimers();
});

it('без выбранной группы открывает граф первой в списке', async () => {
  renderScreen();
  await waitFor(() => expect(lastRequestedGraph()).toBe(farm.id));
  expect(currentLink()).toHaveTextContent('Ферма');
  expect(await screen.findByRole('img', { name: 'Граф группы' })).toBeInTheDocument();
});

it('по клику в списке открывает граф выбранной группы', async () => {
  renderScreen();
  await waitFor(() => expect(lastRequestedGraph()).toBe(farm.id));
  fireEvent.click(groupLink('Кольцо'));
  await waitFor(() => expect(lastRequestedGraph()).toBe(ring.id));
  expect(currentLink()).toHaveTextContent('Кольцо');
});

it('открывает группу из адреса', async () => {
  renderScreen(`/?group=${chain.id}`);
  await waitFor(() => expect(lastRequestedGraph()).toBe(chain.id));
  expect(currentLink()).toHaveTextContent('Цепочка');
});

it('не сбрасывает выбор, когда сверху приходит новая группа', async () => {
  vi.useFakeTimers({ shouldAdvanceTime: true });
  renderScreen(`/?group=${address.id}`);
  await waitForScreen(() => expect(currentLink()).toHaveTextContent('Адрес'));
  respond([{ ...ring, id: '00000000-0000-4000-8000-000000000099' }, farm, address, chain, ring]);
  await advance(3000);
  await waitForScreen(() =>
    expect(screen.getAllByRole('link', { name: /Кольцо/ })).toHaveLength(2)
  );
  expect(currentLink()).toHaveTextContent('Адрес');
  expect(lastRequestedGraph()).toBe(address.id);
});

it('для группы без графа пишет, что граф не найден', async () => {
  const unknown = { ...farm, id: '00000000-0000-4000-8000-000000000042' };
  respond([unknown]);
  renderScreen();
  expect(await screen.findByText('Граф группы не найден')).toBeVisible();
  expect(screen.queryByRole('button', { name: 'Повторить' })).not.toBeInTheDocument();
});
