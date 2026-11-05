import { render, screen, within, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { server } from '@/mocks/server';
import { App } from '@/App';
import type { GroupListResponse } from '@/api/client';

afterEach(() => vi.unstubAllEnvs());

function renderWithClient() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  );
}

describe('Подозрительные группы', () => {
  it('показывает шесть полей группы, сумму в рублях и время UTC', async () => {
    server.use(
      http.get('*/api/v1/groups', ({ request }) => {
        const url = new URL(request.url);
        if (url.searchParams.get('limit') !== '20' || url.searchParams.has('cursor')) {
          return new HttpResponse(null, { status: 400 });
        }
        return HttpResponse.json<GroupListResponse>({
          items: [
            {
              id: '00000000-0000-4000-8000-000000000001',
              scheme: 'farm',
              status: 'new',
              risk_level: 'high',
              risk_score: 87,
              risk_parts: {},
              detected_at: '2026-10-07T10:00:00Z',
              first_event_at: '2026-10-07T10:00:00Z',
              last_event_at: '2026-10-07T15:30:45+05:00',
              features: { accounts_count: 7, total_amount: 12345, span_minutes: 30 },
            },
          ],
          next_cursor: null,
        });
      })
    );
    renderWithClient();
    expect(await screen.findByRole('heading', { name: 'Подозрительные группы' })).toBeVisible();
    const table = await screen.findByRole('table');
    expect(within(table).getAllByRole('columnheader')).toHaveLength(6);
    for (const value of [
      'Ферма',
      'Высокий · 87',
      '7',
      '123,45 ₽',
      '07.10.2026 10:30:45 UTC',
      'Новая',
    ]) {
      expect(within(table).getByText(value)).toBeVisible();
    }
  });
});

it('показывает загрузку до первого ответа', async () => {
  server.use(
    http.get('*/api/v1/groups', async () => {
      await new Promise(() => {});
    })
  );
  renderWithClient();
  expect(await screen.findByRole('status')).toHaveTextContent('Загрузка групп');
  expect(screen.queryByRole('table')).not.toBeInTheDocument();
});

it('объясняет пустой список и предлагает запуск через Swagger без кнопки запуска', async () => {
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json<GroupListResponse>({ items: [], next_cursor: null })
    )
  );
  renderWithClient();
  expect(
    await screen.findByText(
      'Подозрительных групп пока нет. Запустите сценарий через Swagger graph-admin-api'
    )
  ).toBeVisible();
  expect(screen.queryByRole('table')).not.toBeInTheDocument();
  expect(screen.queryByRole('button')).not.toBeInTheDocument();
});

it('сохраняет отличие ошибки от пустого ответа и восстанавливается по Повторить', async () => {
  server.use(http.get('*/api/v1/groups', () => new HttpResponse(null, { status: 500 })));
  renderWithClient();
  expect(await screen.findByRole('alert')).toHaveTextContent('Не удалось загрузить список групп');
  expect(screen.queryByText(/Подозрительных групп пока нет/)).not.toBeInTheDocument();
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json<GroupListResponse>({
        items: [
          {
            id: '00000000-0000-4000-8000-000000000002',
            scheme: 'ring',
            status: 'approved',
            risk_level: 'low',
            risk_score: 15,
            risk_parts: {},
            detected_at: '2026-10-07T10:00:00Z',
            first_event_at: '2026-10-07T10:00:00Z',
            last_event_at: '2026-10-07T10:30:00Z',
            features: { accounts_count: 3, total_amount: 100, span_minutes: 30 },
          },
        ],
        next_cursor: null,
      })
    )
  );
  fireEvent.click(screen.getByRole('button', { name: 'Повторить' }));
  expect(await screen.findByText('Кольцо')).toBeVisible();
  expect(screen.getByText('Подтверждена')).toBeVisible();
  expect(screen.queryByRole('alert')).not.toBeInTheDocument();
});

it('по умолчанию получает демонстрационную группу через MSW', async () => {
  renderWithClient();
  expect(await screen.findByText('Ферма')).toBeVisible();
  expect(screen.getByText('Новая')).toBeVisible();
});

it.each(['empty', 'error', 'loading'])(
  'воспроизводит отдельное состояние моков %s',
  async (state) => {
    vi.stubEnv('VITE_GROUPS_MOCK_STATE', state);
    renderWithClient();
    if (state === 'empty') {
      expect(await screen.findByText(/Подозрительных групп пока нет/)).toBeVisible();
    } else if (state === 'error') {
      expect(await screen.findByRole('alert')).toHaveTextContent(
        'Не удалось загрузить список групп'
      );
    } else {
      expect(await screen.findByRole('status')).toHaveTextContent('Загрузка групп');
    }
  }
);
