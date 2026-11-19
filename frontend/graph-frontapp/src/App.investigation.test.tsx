import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { afterEach, beforeEach, expect, it } from 'vitest';
import type { Group, GroupItem } from '@/api/client';
import { createGroupsDemo } from '@/mocks/groupsDemo';
import { mockGroups } from '@/mocks/groupGraphs';
import { groupsPage } from '@/mocks/groupsPage';
import { server } from '@/mocks/server';
import { renderScreen } from '@/test/renderScreen';

const card = () => within(screen.getByRole('complementary', { name: 'Карточка расследования' }));
const openAddress = () =>
  fireEvent.click(screen.getByRole('link', { name: /^Адрес.*Риск средний/ }));
const farmPath = `*/api/v1/groups/${mockGroups[0].item.id}`;
const farmSolutionPath = `${farmPath}/solution`;
const decisionComment = 'Переводы проверены';
const currentGroup = () => screen.getByRole('link', { current: true });
const loadedGroups = () => within(screen.getByRole('list', { name: 'Список групп' }));
const approve = () => card().getByRole('button', { name: 'Подтвердить' });
const decline = () => card().getByRole('button', { name: 'Отклонить' });
const comment = () => card().getByRole('textbox', { name: 'Комментарий' });

async function openFarm() {
  renderScreen();
  await screen.findByRole('complementary', { name: 'Карточка расследования' });
  expect(await card().findByText('Новая')).toBeVisible();
  return currentGroup();
}

function delaySaveError() {
  let finish = () => {};
  const ready = new Promise<void>((resolve) => {
    finish = resolve;
  });
  server.use(
    http.patch(
      farmSolutionPath,
      async () => {
        await ready;
        return new HttpResponse(null, { status: 500 });
      },
      { once: true }
    )
  );
  return finish;
}

beforeEach(() => {
  server.use(http.get('*/api/v1/groups/:id/graph', () => new HttpResponse(null, { status: 404 })));
});

afterEach(() => {
  server.events.removeAllListeners();
});

it('показывает карточку фермы с текстом и слагаемыми риска по шаблону', async () => {
  renderScreen();
  await screen.findByRole('complementary', { name: 'Карточка расследования' });
  const intro = await card().findByText('Риск: высокий (75/100)');
  expect(intro).toBeVisible();
  expect(card().getByText('Счетов в группе: 4, оборот: 161 000 ₽.')).toBeVisible();
  expect(card().getByRole('heading', { name: 'Что произошло:' })).toBeVisible();
  expect(
    card().getByText(
      'С одного устройства заходили в личный кабинет разные клиенты (всего 3). Деньги с их счетов ушли на общий счёт для вывода — вероятный сборщик. Это признак фермы: возможно, за разными счетами стоит один организатор, который работает с одного устройства и собирает деньги на своём счёте.'
    )
  ).toBeVisible();
  expect(card().getByRole('heading', { name: 'Обоснование риска:' })).toBeVisible();
  expect(
    card()
      .getAllByRole('listitem')
      .map((item) => item.textContent)
  ).toEqual([
    'Тип схемы «ферма»: +50.',
    'Скорость: 28 мин от первого до последнего события — +15.',
    'Клиентов на одном устройстве: 3 — +0.',
    'Есть общий счёт для вывода денег: +10.',
  ]);
  expect(card().getByText('Итого: 75/100 — высокий.')).toBeVisible();
  expect(card().getByText('Новая')).toBeVisible();

  openAddress();
  expect(await card().findByText('Риск: средний (70/100)')).toBeVisible();
  expect(card().getByText('Счетов в группе: 6, оборот: 210 000 ₽.')).toBeVisible();
  expect(card().queryByRole('heading', { name: 'Что произошло:' })).not.toBeInTheDocument();
});

it('объясняет риск фермы без общего счёта и считает уникальных клиентов', async () => {
  const group = createGroupsDemo().getGroup(mockGroups[0].item.id);
  if (!group) throw new Error('Нет моковой фермы');
  server.use(
    http.get(farmPath, () =>
      HttpResponse.json({
        ...group,
        features: { ...group.features, total_amount: 16100001 },
        risk_score: 65,
        risk_level: 'medium',
        logins: [...group.logins, { ...group.logins[0], event_id: 'repeated-login' }],
        risk_parts: { farm: { scheme: 50, speed: 15, clients: 0, collector: 0 } },
      })
    )
  );
  renderScreen();
  await screen.findByRole('complementary', { name: 'Карточка расследования' });
  expect(await card().findByText('Риск: средний (65/100)')).toBeVisible();
  expect(card().getByText('Счетов в группе: 4, оборот: 161 000 ₽.')).toBeVisible();
  expect(
    card().getByText(
      'С одного устройства заходили в личный кабинет разные клиенты (всего 3). Общего счёта для вывода не видно, но несколько клиентов на одном устройстве — уже признак фермы: возможно, счетами управляет один человек.'
    )
  ).toBeVisible();
  expect(card().getByText('Общего счёта для вывода денег нет: +0.')).toBeVisible();
  expect(card().getByText('Итого: 65/100 — средний.')).toBeVisible();
});

it('сообщает об ошибке карточки в её панели и позволяет открыть другую группу', async () => {
  server.use(http.get(farmPath, () => new HttpResponse(null, { status: 500 })));
  renderScreen();
  await screen.findByRole('complementary', { name: 'Карточка расследования' });
  expect(await card().findByRole('alert')).toHaveTextContent(
    'Не удалось загрузить карточку группы'
  );
  expect(screen.getByRole('main')).toBeVisible();
  openAddress();
  expect(await card().findByText('Риск: средний (70/100)')).toBeVisible();
});

it.each([
  ['Подтвердить', 'approved', 'Подтверждена'],
  ['Отклонить', 'declined', 'Отклонена'],
])(
  'действием «%s» меняет статус карточки и списка и сохраняет его после обновления',
  async (action, status, label) => {
    const requests: { path: string; body: unknown }[] = [];
    const farmDetailPath = `/api/v1/groups/${mockGroups[0].item.id}`;
    let listResponses = 0;
    let farmDetailResponses = 0;
    server.events.on('request:start', async ({ request }) => {
      const url = new URL(request.url);
      if (request.method === 'PATCH') {
        requests.push({ path: url.pathname, body: await request.clone().json() });
      }
    });
    server.events.on('response:mocked', ({ request, response }) => {
      if (request.method !== 'GET' || !response.ok) return;
      const path = new URL(request.url).pathname;
      if (path === '/api/v1/groups') listResponses++;
      if (path === farmDetailPath) farmDetailResponses++;
    });
    renderScreen();
    await screen.findByRole('complementary', { name: 'Карточка расследования' });
    expect(await card().findByText('Новая')).toBeVisible();
    const farmLink = currentGroup();
    const actionButton = card().getByRole('button', { name: action });
    const comment = card().getByRole('textbox', { name: 'Комментарий' });
    for (const value of ['', '   ']) {
      fireEvent.change(comment, { target: { value } });
      expect(actionButton).toBeDisabled();
      fireEvent.click(actionButton);
    }
    expect(requests).toEqual([]);
    expect(card().getByText('Новая')).toBeVisible();

    fireEvent.click(screen.getByRole('button', { name: 'Показать ещё' }));
    await waitFor(() => expect(loadedGroups().getAllByRole('listitem')).toHaveLength(29));
    const listResponsesBeforeSave = listResponses;
    fireEvent.change(comment, { target: { value: `  ${decisionComment}  ` } });
    fireEvent.click(actionButton);
    expect(await card().findByText(label)).toBeVisible();
    expect(requests).toEqual([
      {
        path: `${farmDetailPath}/solution`,
        body: { status, decision_comment: decisionComment },
      },
    ]);
    expect(currentGroup()).toBe(farmLink);
    expect(within(farmLink).getByText(label)).toBeVisible();
    expect(card().queryByRole('button', { name: 'Подтвердить' })).not.toBeInTheDocument();
    expect(card().queryByRole('button', { name: 'Отклонить' })).not.toBeInTheDocument();
    expect(loadedGroups().getAllByRole('listitem')).toHaveLength(29);

    openAddress();
    expect(await card().findByText('Риск: средний (70/100)')).toBeVisible();
    await waitFor(() => expect(listResponses).toBeGreaterThanOrEqual(listResponsesBeforeSave + 4), {
      timeout: 5500,
    });
    expect(within(farmLink).getByText(label)).toBeVisible();
    expect(loadedGroups().getAllByRole('listitem')).toHaveLength(29);
    const farmResponsesBeforeReopening = farmDetailResponses;
    fireEvent.click(farmLink);
    await waitFor(() => expect(farmDetailResponses).toBeGreaterThan(farmResponsesBeforeReopening));
    expect(await card().findByText(label)).toBeVisible();
  },
  10000
);

it('блокирует оба решения во время отправки и сохраняет комментарий для повтора после ошибки', async () => {
  const finish = delaySaveError();
  const requests: unknown[] = [];
  server.events.on('request:start', async ({ request }) => {
    if (request.method === 'PATCH') requests.push(await request.clone().json());
  });
  await openFarm();
  fireEvent.change(comment(), { target: { value: decisionComment } });
  fireEvent.click(approve());
  fireEvent.click(decline());
  expect(await card().findByRole('status')).toHaveTextContent('Сохранение решения…');
  expect(approve()).toBeDisabled();
  expect(decline()).toBeDisabled();
  fireEvent.click(decline());
  await waitFor(() => expect(requests).toHaveLength(1));

  finish();
  expect(await card().findByRole('alert')).toHaveTextContent('Не удалось сохранить решение');
  expect(comment()).toHaveValue(decisionComment);
  expect(approve()).toBeEnabled();
  fireEvent.click(approve());
  expect(await card().findByText('Подтверждена')).toBeVisible();
  expect(requests).toEqual([
    { status: 'approved', decision_comment: decisionComment },
    { status: 'approved', decision_comment: decisionComment },
  ]);
});

it('открывает другую группу без черновика и сохраняет блокировку при возврате к группе с отправкой', async () => {
  const finish = delaySaveError();

  const farmLink = await openFarm();
  fireEvent.change(comment(), { target: { value: 'Черновик фермы' } });
  openAddress();
  expect(await card().findByText('Риск: средний (70/100)')).toBeVisible();
  expect(comment()).toHaveValue('');

  fireEvent.click(farmLink);
  expect(await card().findByText('Риск: высокий (75/100)')).toBeVisible();
  expect(comment()).toHaveValue('');
  fireEvent.change(comment(), { target: { value: decisionComment } });
  fireEvent.click(approve());
  expect(await card().findByRole('status')).toHaveTextContent('Сохранение решения…');

  openAddress();
  expect(await card().findByText('Риск: средний (70/100)')).toBeVisible();
  expect(comment()).toBeEnabled();
  fireEvent.click(farmLink);
  expect(await card().findByRole('status')).toHaveTextContent('Сохранение решения…');
  expect(approve()).toBeDisabled();
  expect(decline()).toBeDisabled();
  finish();
  await waitFor(() => expect(comment()).toBeEnabled());
});

it('при конфликте перечитывает статус карточки и списка с сервера', async () => {
  const demo = createGroupsDemo();
  const groupId = mockGroups[0].item.id;
  let status: GroupItem['status'] = 'new';
  server.use(
    http.patch(farmSolutionPath, () => {
      status = 'declined';
      return new HttpResponse(null, { status: 409 });
    }),
    http.get(farmPath, () => {
      const group = demo.getGroup(groupId);
      return group ? HttpResponse.json<Group>({ ...group, status }) : HttpResponse.error();
    }),
    http.get('*/api/v1/groups', ({ request }) =>
      HttpResponse.json(
        groupsPage(
          demo.getItems().map((item) => (item.id === groupId ? { ...item, status } : item)),
          request
        )
      )
    )
  );

  const farmLink = await openFarm();
  fireEvent.change(comment(), { target: { value: decisionComment } });
  fireEvent.click(approve());
  expect(await card().findByText('Отклонена')).toBeVisible();
  expect(await within(farmLink).findByText('Отклонена')).toBeVisible();
  expect(card().queryByRole('textbox', { name: 'Комментарий' })).not.toBeInTheDocument();
});
