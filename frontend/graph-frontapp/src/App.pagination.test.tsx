import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { renderScreen } from '@/test/renderScreen';
import { server } from '@/mocks/server';
import type { GroupItem, GroupListResponse } from '@/api/client';
import { groupItem } from '@/test/groupItem';

const items: GroupItem[] = Array.from({ length: 25 }, (_, index) => ({
  ...groupItem,
  id: `00000000-0000-4000-8000-${String(index + 1).padStart(12, '0')}`,
  risk_score: index + 1,
  features: { ...groupItem.features, accounts_count: index + 1 },
}));
const cursor = 'opaque/+next-page==';

it('добавляет следующие группы по непрозрачному курсору, сохраняя таблицу во время загрузки', async () => {
  let finishPage!: () => void;
  const pendingPage = new Promise<void>((resolve) => {
    finishPage = resolve;
  });
  server.use(
    http.get('*/api/v1/groups', async ({ request }) => {
      const query = new URL(request.url).searchParams;
      if (query.get('limit') !== '20') return new HttpResponse(null, { status: 400 });
      if (!query.has('cursor'))
        return HttpResponse.json<GroupListResponse>({
          items: items.slice(0, 20),
          next_cursor: cursor,
        });
      if (query.get('cursor') !== cursor) return new HttpResponse(null, { status: 400 });
      await pendingPage;
      return HttpResponse.json<GroupListResponse>({ items: items.slice(20), next_cursor: null });
    })
  );
  renderScreen();
  const list = await screen.findByRole('list', { name: 'Список групп' });
  expect(within(list).getAllByRole('listitem')).toHaveLength(20);
  fireEvent.click(screen.getByRole('button', { name: 'Показать ещё' }));
  await waitFor(() => expect(screen.getByRole('button', { name: 'Показать ещё' })).toBeDisabled());
  expect(within(list).getAllByRole('listitem')).toHaveLength(20);
  finishPage();
  await waitFor(() => expect(within(list).getAllByRole('listitem')).toHaveLength(25));
  expect(screen.queryByRole('button', { name: 'Показать ещё' })).not.toBeInTheDocument();
});

it('сохраняет загруженные группы после ошибки следующей страницы и повторяет её без дублей', async () => {
  let failNextPage = true;
  server.use(
    http.get('*/api/v1/groups', ({ request }) => {
      const query = new URL(request.url).searchParams;
      if (!query.has('cursor'))
        return HttpResponse.json<GroupListResponse>({
          items: items.slice(0, 20),
          next_cursor: cursor,
        });
      if (query.get('cursor') !== cursor || failNextPage)
        return HttpResponse.json({ detail: 'Курсор страницы недействителен' }, { status: 400 });
      return HttpResponse.json<GroupListResponse>({ items: items.slice(20), next_cursor: null });
    })
  );
  renderScreen();
  const list = await screen.findByRole('list', { name: 'Список групп' });
  fireEvent.click(screen.getByRole('button', { name: 'Показать ещё' }));
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Не удалось загрузить следующую страницу'
  );
  expect(screen.getByRole('alert')).toHaveTextContent('Курсор страницы недействителен');
  expect(within(list).getAllByRole('listitem')).toHaveLength(20);
  failNextPage = false;
  fireEvent.click(screen.getByRole('button', { name: 'Повторить загрузку' }));
  await waitFor(() => expect(within(list).getAllByRole('listitem')).toHaveLength(25));
  expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  expect(screen.queryByRole('button')).not.toBeInTheDocument();
});
