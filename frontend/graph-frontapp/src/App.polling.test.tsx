import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { advance, renderScreen, waitForScreen } from '@/test/renderScreen';
import { groupItem } from '@/test/groupItem';
import { server } from '@/mocks/server';
import type { GroupItem, GroupListResponse } from '@/api/client';

const group = { ...groupItem, risk_score: 87 };

beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

async function openScreen() {
  const view = renderScreen();
  await waitForScreen(() =>
    expect(screen.getByRole('list', { name: 'Список групп' })).toBeVisible()
  );
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
  expect(screen.getByLabelText('Риск высокий: 87')).toBeVisible();
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
  await advance(2000);
  expect(screen.queryByText('Кольцо')).not.toBeInTheDocument();
  await advance(1000);
  await waitForScreen(() => expect(screen.getByLabelText('Риск высокий: 95')).toBeVisible());
  const rows = within(screen.getByRole('list', { name: 'Список групп' })).getAllByRole('listitem');
  expect(rows).toHaveLength(2);
  expect(rows[0]).toHaveTextContent('Кольцо');
  expect(within(rows[1]).getByLabelText('Риск высокий: 95')).toBeVisible();
  expect(rows[1]).toHaveTextContent('678,90');
  expect(rows[1]).toHaveTextContent('07.10.2026 11:00');
  expect(rows[1]).toHaveTextContent('Подтверждена');
});

it('сохраняет таблицу при ошибке обновления и восстанавливается следующим опросом', async () => {
  respond([group]);
  await openScreen();
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json({ detail: 'Обновление временно недоступно' }, { status: 503 })
    )
  );
  await advance(3000);
  await waitForScreen(() =>
    expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список')
  );
  expect(screen.getByRole('alert')).toHaveTextContent('Обновление временно недоступно');
  expect(
    within(screen.getByRole('list', { name: 'Список групп' })).getByLabelText('Риск высокий: 87')
  ).toBeVisible();
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
  respond([{ ...group, risk_score: 95 }]);
  await advance(3000);
  await waitForScreen(() => expect(screen.getByLabelText('Риск высокий: 95')).toBeVisible());
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
  expect(screen.getByLabelText('Риск высокий: 87')).toBeVisible();
  visibility.mockReturnValue('visible');
  fireEvent(document, new Event('visibilitychange', { bubbles: true }));
  await advance(1);
  await waitForScreen(() => expect(screen.getByLabelText('Риск высокий: 95')).toBeVisible());
});

it('сохраняет таблицу пока ответ фонового обновления задерживается', async () => {
  respond([group]);
  await openScreen();
  let finish!: () => void;
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
  await waitForScreen(() => expect(finish).not.toEqual(undefined));
  expect(
    within(screen.getByRole('list', { name: 'Список групп' })).getByLabelText('Риск высокий: 87')
  ).toBeVisible();
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
  finish();
  await advance(1);
  await waitForScreen(() => expect(screen.getByLabelText('Риск высокий: 95')).toBeVisible());
});
