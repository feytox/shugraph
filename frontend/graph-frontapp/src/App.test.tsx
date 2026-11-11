import { screen, within, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/mocks/server';
import { renderScreen } from '@/test/renderScreen';
import type { GroupListResponse } from '@/api/client';
import { groupItem } from '@/test/groupItem';

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
              ...groupItem,
              risk_score: 87,
              last_event_at: '2026-10-07T15:30:45+05:00',
              features: { accounts_count: 7, total_amount: 12345, span_minutes: 30 },
            },
          ],
          next_cursor: null,
        });
      })
    );
    renderScreen();
    expect(await screen.findByRole('heading', { name: 'Подозрительные группы' })).toBeVisible();
    const [item] = within(await screen.findByRole('list', { name: 'Список групп' })).getAllByRole(
      'listitem'
    );
    for (const value of ['Ферма', '7 счетов', '123,45 ₽', '07.10.2026 10:30', 'Новая']) {
      expect(item).toHaveTextContent(value);
    }
    expect(within(item).getByLabelText('Риск высокий: 87')).toBeVisible();
  });
});

it('показывает загрузку до первого ответа', async () => {
  server.use(
    http.get('*/api/v1/groups', async () => {
      await new Promise(() => {});
    })
  );
  renderScreen();
  expect(await screen.findByRole('status')).toHaveTextContent('Загрузка групп');
  expect(screen.queryByRole('list', { name: 'Список групп' })).not.toBeInTheDocument();
});

it('объясняет пустой список и предлагает запуск через Swagger без кнопки запуска', async () => {
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json<GroupListResponse>({ items: [], next_cursor: null })
    )
  );
  renderScreen();
  expect(
    await screen.findByText(
      'Подозрительных групп пока нет. Запустите сценарий через Swagger graph-admin-api'
    )
  ).toBeVisible();
  expect(screen.queryByRole('list', { name: 'Список групп' })).not.toBeInTheDocument();
  expect(screen.queryByRole('button', { name: /сценари/i })).not.toBeInTheDocument();
});

it('сохраняет отличие ошибки от пустого ответа и восстанавливается по Повторить', async () => {
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json({ detail: 'Хранилище групп временно недоступно' }, { status: 503 })
    )
  );
  renderScreen();
  expect(await screen.findByRole('alert')).toHaveTextContent('Хранилище групп временно недоступно');
  expect(screen.queryByText(/Подозрительных групп пока нет/)).not.toBeInTheDocument();
  server.use(
    http.get('*/api/v1/groups', () =>
      HttpResponse.json<GroupListResponse>({
        items: [
          {
            ...groupItem,
            id: '00000000-0000-4000-8000-000000000002',
            scheme: 'ring',
            status: 'approved',
            risk_level: 'low',
            risk_score: 15,
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
