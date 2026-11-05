import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { App } from '@/App';
import { server } from '@/mocks/server';
import { groupsPage } from '@/mocks/groupsPage';
import { createGroupsDemo } from '@/mocks/groupsDemo';
import type { GroupItem } from '@/api/client';
import { groupItem } from '@/test/groupItem';

const original: GroupItem[] = Array.from({ length: 65 }, (_, index) => ({
  ...groupItem,
  id: `group-${index + 1}`,
  risk_score: index + 1,
  features: { ...groupItem.features, accounts_count: index + 1 },
}));

beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});
async function advance(ms = 1) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(ms);
  });
  for (let flush = 0; flush < 6; flush++)
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1);
    });
}
async function openScreen() {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <App />
    </QueryClientProvider>
  );
  await advance();
}
function counts() {
  return within(screen.getByRole('table'))
    .getAllByRole('row')
    .slice(1)
    .map((row) => within(row).getAllByRole('cell')[2].textContent);
}
async function more() {
  fireEvent.click(screen.getByRole('button', { name: 'Показать ещё' }));
  await advance();
}

it('сохраняет загруженные группы при поступлении сверху, обновляет дальние строки и продолжает курсор', async () => {
  let items = original;
  server.use(
    http.get('*/api/v1/groups', ({ request }) => HttpResponse.json(groupsPage(items, request)))
  );
  await openScreen();
  await more();
  expect(counts()).toEqual(Array.from({ length: 40 }, (_, i) => String(i + 1)));
  items = [
    { ...original[0], id: 'arrival', features: { ...original[0].features, accounts_count: 100 } },
    ...original.map((item, index) =>
      index === 39
        ? {
            ...item,
            status: 'approved' as const,
            risk_score: 99,
            last_event_at: '2026-10-07T11:00:00Z',
            features: { ...item.features, total_amount: 67890 },
          }
        : item
    ),
  ];
  await advance(3000);
  expect(counts()).toEqual(['100', ...Array.from({ length: 40 }, (_, i) => String(i + 1))]);
  const last = within(screen.getByRole('table')).getAllByRole('row').at(-1)!;
  expect(last).toHaveTextContent('Высокий · 99');
  expect(last).toHaveTextContent('678,90');
  expect(last).toHaveTextContent('07.10.2026 11:00:00 UTC');
  expect(last).toHaveTextContent('Подтверждена');
  await more();
  expect(counts()).toEqual(['100', ...Array.from({ length: 60 }, (_, i) => String(i + 1))]);
});

it.each([
  { fail: false, elapsed: 100 },
  { fail: true, elapsed: 100 },
  { fail: false, elapsed: 3000 },
  { fail: true, elapsed: 3000 },
])(
  'обновляет список сразу после дозагрузки при возврате на вкладку (ошибка: $fail, задержка: $elapsed)',
  async ({ fail, elapsed }) => {
    let items = original;
    let finish!: () => void;
    let pending = false;
    server.use(
      http.get('*/api/v1/groups', async ({ request }) => {
        const result = groupsPage(items, request);
        if (pending && new URL(request.url).searchParams.has('cursor')) {
          await new Promise<void>((resolve) => {
            finish = resolve;
          });
          pending = false;
          if (fail) return new HttpResponse(null, { status: 500 });
        }
        return HttpResponse.json(result);
      })
    );
    await openScreen();
    await more();
    pending = true;
    await more();
    const visibility = vi.spyOn(document, 'visibilityState', 'get');
    visibility.mockReturnValue('hidden');
    fireEvent(document, new Event('visibilitychange', { bubbles: true }));
    items = [
      { ...original[0], id: 'arrival', features: { ...original[0].features, accounts_count: 100 } },
      ...original,
    ];
    await advance(elapsed);
    visibility.mockReturnValue('visible');
    fireEvent(document, new Event('visibilitychange', { bubbles: true }));
    await advance();
    expect(counts()).toEqual(Array.from({ length: 40 }, (_, i) => String(i + 1)));
    expect(screen.getByRole('button')).toBeDisabled();
    finish();
    await advance();
    expect(counts()).toEqual([
      '100',
      ...Array.from({ length: fail ? 40 : 60 }, (_, i) => String(i + 1)),
    ]);
    if (fail) {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Не удалось загрузить следующую страницу'
      );
      fireEvent.click(screen.getByRole('button', { name: 'Повторить загрузку' }));
      await advance();
      expect(counts()).toHaveLength(61);
      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    }
  }
);

it('сохраняет все страницы при ошибке обновления и снимает ошибку только после успешного обновления', async () => {
  let fail = false;
  server.use(
    http.get('*/api/v1/groups', ({ request }) =>
      fail
        ? new HttpResponse(null, { status: 500 })
        : HttpResponse.json(groupsPage(original, request))
    )
  );
  await openScreen();
  await more();
  fail = true;
  await advance(3000);
  expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список');
  expect(counts()).toHaveLength(40);
  fail = false;
  await advance(3000);
  expect(counts()).toHaveLength(40);
  expect(screen.queryByRole('alert')).not.toBeInTheDocument();
});

it('успех дозагрузки сохраняет ошибку обновления до успешного обновления', async () => {
  let failRefresh = false;

  server.use(
    http.get('*/api/v1/groups', ({ request }) => {
      const cursor = new URL(request.url).searchParams.get('cursor');
      if (!cursor && failRefresh) {
        return new HttpResponse(null, { status: 500 });
      }
      return HttpResponse.json(groupsPage(original, request));
    })
  );
  await openScreen();
  await more();
  failRefresh = true;
  await advance(3000);
  expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список');
  await more();
  expect(counts()).toHaveLength(60);
  expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список');
  failRefresh = false;
  await advance(3000);
  expect(screen.queryByRole('alert')).not.toBeInTheDocument();
});

it('после дозагрузки показывает полный поток 25→26→27→28 и начинает заново после сброса демо', async () => {
  let demo = createGroupsDemo();
  server.use(
    http.get('*/api/v1/groups', ({ request }) =>
      HttpResponse.json(groupsPage(demo.getItems(), request))
    )
  );
  const view = render(
    <QueryClientProvider client={new QueryClient()}>
      <App />
    </QueryClientProvider>
  );
  await advance();
  await more();
  expect(counts()).toHaveLength(25);
  for (const size of [26, 27, 28]) {
    for (let tick = 0; tick < 4; tick++) await advance(3000);
    expect(counts()).toHaveLength(size);
    expect(new Set(counts()).size).toBe(size);
    expect(counts()[0]).toBe(String(size));
  }
  view.unmount();
  demo = createGroupsDemo();
  await openScreen();
  await more();
  expect(counts()).toHaveLength(25);
  expect(counts()[0]).toBe('25');
});

it.each([false, true])(
  'сохраняет таблицу и блокирует дозагрузку пока обновление задержано (ошибка: %s)',
  async (fail) => {
    let finish!: () => void;
    let pending = false;
    server.use(
      http.get('*/api/v1/groups', async ({ request }) => {
        if (pending && !new URL(request.url).searchParams.has('cursor')) {
          await new Promise<void>((resolve) => {
            finish = resolve;
          });
          pending = false;
          if (fail) return new HttpResponse(null, { status: 500 });
        }
        return HttpResponse.json(groupsPage(original, request));
      })
    );
    await openScreen();
    await more();
    pending = true;
    await advance(3000);
    expect(counts()).toHaveLength(40);
    expect(screen.getByRole('button')).toBeDisabled();
    fireEvent.click(screen.getByRole('button'));
    finish();
    await advance();
    expect(counts()).toHaveLength(40);
    await more();
    expect(counts()).toHaveLength(60);
    if (fail) expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список');
    else expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  }
);

it('после пустого списка загружает полную первую порцию при появлении групп', async () => {
  let items: GroupItem[] = [];
  server.use(
    http.get('*/api/v1/groups', ({ request }) => HttpResponse.json(groupsPage(items, request)))
  );
  await openScreen();
  expect(screen.getByText(/Подозрительных групп пока нет/)).toBeVisible();
  items = original.slice(0, 25);
  await advance(3000);
  expect(counts()).toEqual(Array.from({ length: 20 }, (_, i) => String(i + 1)));
  await more();
  expect(counts()).toHaveLength(25);
});
