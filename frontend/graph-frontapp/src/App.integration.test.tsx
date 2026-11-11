import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { advance, renderScreen, waitForScreen } from '@/test/renderScreen';
import { server } from '@/mocks/server';
import { groupsPage } from '@/mocks/groupsPage';
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
async function openScreen() {
  renderScreen();
  await waitForScreen(() =>
    expect(screen.getByRole('list', { name: 'Список групп' })).toBeVisible()
  );
}
function counts() {
  return within(screen.getByRole('list', { name: 'Список групп' }))
    .getAllByRole('listitem')
    .map(
      (item) =>
        within(item)
          .getByText(/сч[её]т/)
          .textContent!.split(' ')[0]
    );
}
async function more(size: number) {
  fireEvent.click(screen.getByRole('button', { name: 'Показать ещё' }));
  await waitForScreen(() => expect(counts()).toHaveLength(size));
}

it('обновляет загруженные страницы со свежими курсорами и продолжает без пропусков', async () => {
  let items = original;
  server.use(
    http.get('*/api/v1/groups', ({ request }) => HttpResponse.json(groupsPage(items, request)))
  );
  await openScreen();
  await more(40);
  expect(counts()).toEqual(Array.from({ length: 40 }, (_, i) => String(i + 1)));
  items = [
    { ...original[0], id: 'arrival', features: { ...original[0].features, accounts_count: 100 } },
    ...original.map((item, index) =>
      index === 38
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
  await waitForScreen(() =>
    expect(counts()).toEqual(['100', ...Array.from({ length: 39 }, (_, i) => String(i + 1))])
  );
  const last = within(screen.getByRole('list', { name: 'Список групп' }))
    .getAllByRole('listitem')
    .at(-1)!;
  expect(within(last).getByLabelText('Риск высокий: 99')).toBeVisible();
  expect(last).toHaveTextContent('678,90');
  expect(last).toHaveTextContent('07.10.2026 11:00');
  expect(last).toHaveTextContent('Подтверждена');
  await more(60);
  expect(counts()).toEqual(['100', ...Array.from({ length: 59 }, (_, i) => String(i + 1))]);
});

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
  await more(40);
  fail = true;
  await advance(3000);
  await waitForScreen(() =>
    expect(screen.getByRole('alert')).toHaveTextContent('Не удалось обновить список')
  );
  expect(counts()).toHaveLength(40);
  fail = false;
  await advance(3000);
  expect(counts()).toHaveLength(40);
  await waitForScreen(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument());
});

it('после пустого списка загружает полную первую порцию при появлении групп', async () => {
  let items: GroupItem[] = [];
  server.use(
    http.get('*/api/v1/groups', ({ request }) => HttpResponse.json(groupsPage(items, request)))
  );
  renderScreen();
  await waitForScreen(() =>
    expect(screen.getByText(/Подозрительных групп пока нет/)).toBeVisible()
  );
  items = original.slice(0, 25);
  await advance(3000);
  await waitForScreen(() =>
    expect(counts()).toEqual(Array.from({ length: 20 }, (_, i) => String(i + 1)))
  );
  await more(25);
  expect(counts()).toHaveLength(25);
});

it('после удаления группы обновляет только загруженное число страниц', async () => {
  let items = original;
  const cursors: (string | null)[] = [];
  server.use(
    http.get('*/api/v1/groups', ({ request }) => {
      cursors.push(new URL(request.url).searchParams.get('cursor'));
      return HttpResponse.json(groupsPage(items, request));
    })
  );
  await openScreen();
  await more(40);
  cursors.length = 0;
  items = original.slice(1);
  await advance(3000);
  await waitForScreen(() => expect(counts()[0]).toBe('2'));
  expect(counts()).toHaveLength(40);
  expect(cursors).toEqual([null, btoa('group-21')]);
});

it('убирает дубли между страницами', async () => {
  server.use(
    http.get('*/api/v1/groups', ({ request }) =>
      HttpResponse.json(
        new URL(request.url).searchParams.has('cursor')
          ? { items: original.slice(19, 25), next_cursor: null }
          : { items: original.slice(0, 20), next_cursor: 'next' }
      )
    )
  );
  await openScreen();
  await more(25);
  expect(new Set(counts()).size).toBe(25);
});

it('позволяет дозагрузить страницу во время фонового обновления', async () => {
  let pending = false;
  let finish!: () => void;
  server.use(
    http.get('*/api/v1/groups', async ({ request }) => {
      if (pending && !new URL(request.url).searchParams.has('cursor')) {
        await new Promise<void>((resolve) => {
          finish = resolve;
        });
      }
      return HttpResponse.json(groupsPage(original, request));
    })
  );
  await openScreen();
  await more(40);
  pending = true;
  await advance(3000);
  await waitForScreen(() => expect(finish).toBeTypeOf('function'));
  expect(screen.getByRole('button', { name: 'Показать ещё' })).toBeEnabled();
  await more(60);
  pending = false;
  finish();
  await advance(100);
  expect(counts()).toHaveLength(60);
});

it('после ошибки первой загрузки ждёт явного повтора', async () => {
  let requests = 0;
  server.use(
    http.get('*/api/v1/groups', () => {
      requests++;
      return new HttpResponse(null, { status: 500 });
    })
  );
  renderScreen();
  await waitForScreen(() => expect(screen.getByRole('alert')).toHaveTextContent('Ошибка загрузки'));
  await advance(9000);
  expect(requests).toBe(1);
  expect(screen.getByRole('button', { name: 'Повторить' })).toBeEnabled();
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
});
