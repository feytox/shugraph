import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { App } from '@/App';
import { server } from '@/mocks/server';
import type { GroupItem, GroupListResponse } from '@/api/client';
import { createGroupsDemo } from '@/mocks/groupsDemo';

const group: GroupItem = {
  id: '00000000-0000-4000-8000-000000000001',
  scheme: 'farm',
  status: 'new',
  risk_level: 'high',
  risk_score: 87,
  risk_parts: {},
  detected_at: '2026-10-07T10:00:00Z',
  first_event_at: '2026-10-07T10:00:00Z',
  last_event_at: '2026-10-07T10:30:45Z',
  features: { accounts_count: 7, total_amount: 12345, span_minutes: 30 },
};

beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  vi.unstubAllEnvs();
});

async function advance(milliseconds: number) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(milliseconds);
  });
  // Allow the HTTP response and Query's notification to reach the rendered screen.
  await act(async () => {
    await vi.advanceTimersByTimeAsync(1);
  });
}

async function openScreen() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const view = render(
    <QueryClientProvider client={client}>
      <App />
    </QueryClientProvider>
  );
  await advance(1);
  return view;
}

function respond(items: GroupItem[]) {
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json<GroupListResponse>({ items, next_cursor: null })
    )
  );
}

it('обновляет строки через 3 секунды и добавляет новую группу сверху без дублей', async () => {
  respond([group]);
  await openScreen();
  expect(screen.getByText('Высокий · 87')).toBeVisible();
  respond([
    { ...group, id: '00000000-0000-4000-8000-000000000002', scheme: 'ring' },
    {
      ...group,
      risk_score: 95,
      status: 'approved',
      last_event_at: '2026-10-07T11:00:00Z',
      features: { ...group.features, total_amount: 67890 },
    },
  ]);
  await advance(2990);
  expect(screen.queryByText('Кольцо')).not.toBeInTheDocument();
  await advance(10);
  const rows = within(screen.getByRole('table')).getAllByRole('row');
  expect(rows).toHaveLength(3);
  expect(rows[1]).toHaveTextContent('Кольцо');
  expect(rows[2]).toHaveTextContent('Высокий · 95');
  expect(rows[2]).toHaveTextContent('678,90');
  expect(rows[2]).toHaveTextContent('07.10.2026 11:00:00 UTC');
  expect(rows[2]).toHaveTextContent('Подтверждена');
});

it('сохраняет таблицу при ошибке обновления и восстанавливается следующим опросом', async () => {
  respond([group]);
  await openScreen();
  server.use(http.get('*/api/v1/groups', () => new HttpResponse(null, { status: 500 })));
  await advance(3000);
  expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список');
  expect(screen.getByRole('table')).toHaveTextContent('Высокий · 87');
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
  respond([{ ...group, risk_score: 95 }]);
  await advance(3000);
  expect(screen.getByText('Высокий · 95')).toBeVisible();
  expect(screen.queryByRole('alert')).not.toBeInTheDocument();
});

it('не опрашивает скрытую вкладку и обновляет сразу при возвращении', async () => {
  respond([group]);
  await openScreen();
  const visibility = vi.spyOn(document, 'visibilityState', 'get');
  visibility.mockReturnValue('hidden');
  fireEvent(document, new Event('visibilitychange', { bubbles: true }));
  let requests = 0;
  server.use(
    http.get('*/api/v1/groups', () => {
      requests++;
      return HttpResponse.json<GroupListResponse>({
        items: [{ ...group, risk_score: 95 }],
        next_cursor: null,
      });
    })
  );
  for (let tick = 0; tick < 3; tick++) await advance(3000);
  expect(requests).toBe(0);
  expect(screen.getByText('Высокий · 87')).toBeVisible();
  visibility.mockReturnValue('visible');
  fireEvent(document, new Event('visibilitychange', { bubbles: true }));
  await advance(1);
  expect(screen.getByText('Высокий · 95')).toBeVisible();
});

it('сохраняет таблицу пока ответ фонового обновления задерживается', async () => {
  respond([group]);
  await openScreen();
  let finish: () => void = () => {};
  server.use(
    http.get('*/api/v1/groups', async () => {
      await new Promise<void>((resolve) => {
        finish = resolve;
      });
      return HttpResponse.json<GroupListResponse>({
        items: [{ ...group, risk_score: 95 }],
        next_cursor: null,
      });
    })
  );
  await advance(3000);
  expect(screen.getByRole('table')).toHaveTextContent('Высокий · 87');
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
  finish();
  await advance(1);
  expect(screen.getByText('Высокий · 95')).toBeVisible();
});

it('воспроизводит поступления через 10, 20, 30 секунд и обновляет прежнюю строку', async () => {
  let demo = createGroupsDemo();
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json<GroupListResponse>({
        items: demo.getItems().slice(0, 20),
        next_cursor: null,
      })
    )
  );
  const view = await openScreen();
  expect(within(screen.getByRole('table')).getAllByRole('row')).toHaveLength(21);
  expect(screen.getByText('Высокий · 87')).toBeVisible();
  for (let tick = 0; tick < 3; tick++) await advance(3000);
  expect(screen.queryByText('26')).not.toBeInTheDocument();
  await advance(3000);
  expect(screen.getByText('26')).toBeVisible();
  expect(screen.getByText('Высокий · 88')).toBeVisible();
  expect(screen.getByText('223,45 ₽')).toBeVisible();
  expect(screen.getByText('07.10.2026 10:30:46 UTC')).toBeVisible();
  for (let tick = 0; tick < 3; tick++) await advance(3000);
  expect(screen.getByText('27')).toBeVisible();
  expect(screen.getByText('Высокий · 89')).toBeVisible();
  for (let tick = 0; tick < 3; tick++) await advance(3000);
  expect(screen.getByText('28')).toBeVisible();
  expect(screen.getByText('Высокий · 90')).toBeVisible();
  expect(within(screen.getByRole('table')).getAllByRole('row')).toHaveLength(21);
  view.unmount();
  demo = createGroupsDemo();
  await openScreen();
  expect(screen.getByText('Высокий · 87')).toBeVisible();
  expect(screen.queryByText('26')).not.toBeInTheDocument();
});

it('настройка polling-error воспроизводит ошибку обновления и восстановление', async () => {
  vi.stubEnv('VITE_GROUPS_MOCK_STATE', 'polling-error');
  await openScreen();
  expect(screen.getByRole('table')).toBeVisible();
  await advance(3000);
  expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список');
  expect(screen.getByRole('table')).toBeVisible();
  await advance(3000);
  expect(screen.queryByRole('alert')).not.toBeInTheDocument();
});
