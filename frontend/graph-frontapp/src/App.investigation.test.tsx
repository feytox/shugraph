import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { apiClient, type ErrorModel, type Group, type GroupItem } from '@/api/client';
import { createGroupsDemo } from '@/mocks/groupsDemo';
import { mockGroups } from '@/mocks/groupGraphs';
import { groupsPage } from '@/mocks/groupsPage';
import { server } from '@/mocks/server';
import { advance, renderScreen, waitForScreen } from '@/test/renderScreen';

const card = () => within(screen.getByRole('complementary', { name: 'Карточка расследования' }));
const openAddress = () =>
  fireEvent.click(screen.getByRole('link', { name: /^Адрес.*Риск средний/ }));
const farm = mockGroups.find(({ item }) => item.scheme === 'farm');
if (!farm) throw new Error('Нет моковой фермы');
const farmDetailPath = `/api/v1/groups/${farm.item.id}`;
const farmPath = `*${farmDetailPath}`;
const farmSolutionPath = `${farmPath}/solution`;
const evolvingGroupId = '00000000-0000-4000-8000-000000000025';
const evolvingGroupUrl = `/?group=${evolvingGroupId}`;
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

function delaySaveError(detail?: string) {
  let finish = () => {};
  const ready = new Promise<void>((resolve) => {
    finish = resolve;
  });
  server.use(
    http.patch(
      farmSolutionPath,
      async () => {
        await ready;
        return detail ? problem(500, detail) : new HttpResponse(null, { status: 500 });
      },
      { once: true }
    )
  );
  return finish;
}

function problem(status: number, detail: string) {
  return HttpResponse.json<ErrorModel>(
    { type: 'about:blank', status, detail },
    { status, headers: { 'Content-Type': 'application/problem+json' } }
  );
}

function startDemoClock() {
  vi.useFakeTimers();
  vi.setSystemTime(new Date('2026-10-07T11:00:00Z'));
}

async function nextDemoStep() {
  vi.setSystemTime(Date.now() + 10000);
  await advance(3000);
  await waitForScreen(() =>
    expect(loadedGroups().getByLabelText('Риск высокий: 86')).toBeVisible()
  );
}

beforeEach(() => {
  server.use(http.get('*/api/v1/groups/:id/graph', () => new HttpResponse(null, { status: 404 })));
});

afterEach(() => {
  server.events.removeAllListeners();
  vi.useRealTimers();
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
  const group = createGroupsDemo().getGroup(farm.item.id);
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

it('показывает причину ошибки API и позволяет повторно загрузить карточку', async () => {
  const detail = 'Хранилище временно недоступно. Повторите запрос через несколько секунд.';
  server.use(http.get(farmPath, () => problem(503, detail), { once: true }));
  renderScreen();
  await screen.findByRole('complementary', { name: 'Карточка расследования' });
  expect(await card().findByRole('alert')).toHaveTextContent(detail);
  fireEvent.click(card().getByRole('button', { name: /^Повторить/ }));
  expect(await card().findByText('Риск: высокий (75/100)')).toBeVisible();
});

it('показывает причину отсутствия карточки без кнопки повтора', async () => {
  const detail = 'Группа удалена';
  server.use(http.get(farmPath, () => problem(404, detail)));
  renderScreen();
  await screen.findByRole('complementary', { name: 'Карточка расследования' });
  expect(await card().findByText(detail)).toBeVisible();
  expect(card().queryByRole('button', { name: /^Повторить/ })).not.toBeInTheDocument();
});

it('после опроса списка обновляет открытую карточку и сохраняет введённый комментарий', async () => {
  startDemoClock();
  renderScreen(evolvingGroupUrl);
  await waitForScreen(() => expect(card().getByText('Риск: высокий (87/100)')).toBeVisible());
  fireEvent.change(comment(), { target: { value: decisionComment } });

  await nextDemoStep();
  expect(within(currentGroup()).getByLabelText('Риск высокий: 88')).toBeVisible();
  await waitForScreen(() => expect(card().getByText('Риск: высокий (88/100)')).toBeVisible());
  expect(card().getByText('Счетов в группе: 25, оборот: 223,45 ₽.')).toBeVisible();
  expect(comment()).toHaveValue(decisionComment);
});

it('сохраняет риск, оборот и время события решённой демогруппы на следующем шаге потока', async () => {
  startDemoClock();
  const view = renderScreen(evolvingGroupUrl);
  await waitForScreen(() => expect(approve()).toBeDisabled());
  fireEvent.change(comment(), { target: { value: decisionComment } });
  fireEvent.click(approve());
  await waitForScreen(() => expect(card().getByText('Подтверждена')).toBeVisible());

  await nextDemoStep();
  const row = within(currentGroup());
  expect(row.getByLabelText('Риск высокий: 87')).toBeVisible();
  expect(row.getByText(/123,45/)).toBeVisible();

  view.unmount();
  renderScreen(evolvingGroupUrl);
  await waitForScreen(() => expect(card().getByText('Риск: высокий (87/100)')).toBeVisible());
  expect(card().getByText('Счетов в группе: 25, оборот: 123,45 ₽.')).toBeVisible();
  expect(card().getByText('Подтверждена')).toBeVisible();
  const { data } = await apiClient.GET('/api/v1/groups/{group_id}', {
    params: { path: { group_id: evolvingGroupId } },
  });
  expect(data).toMatchObject({ last_event_at: '2026-10-07T10:30:45Z', is_frozen: true });
});

it.each([
  ['Подтвердить', 'approved', 'Подтверждена'],
  ['Отклонить', 'declined', 'Отклонена'],
])(
  'действием «%s» меняет статус карточки и списка и сохраняет его после обновления',
  async (action, status, label) => {
    const requests: { path: string; body: unknown }[] = [];
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
  const detail = 'Сохранение решений временно приостановлено';
  const finish = delaySaveError(detail);
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
  expect(await card().findByRole('alert')).toHaveTextContent(detail);
  expect(comment()).toHaveValue(decisionComment);
  expect(approve()).toBeEnabled();
  fireEvent.click(approve());
  expect(await card().findByText('Подтверждена')).toBeVisible();
  expect(requests).toEqual([
    { status: 'approved', decision_comment: decisionComment },
    { status: 'approved', decision_comment: decisionComment },
  ]);
});

it('сбрасывает черновик при переходе и восстанавливает отправленный комментарий после ошибки', async () => {
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
  expect(await card().findByRole('alert')).toHaveTextContent('Не удалось сохранить решение');
  expect(comment()).toHaveValue(decisionComment);
  expect(approve()).toBeEnabled();
});

it('после опроса обновляет риск, оборот и статус при прежнем времени последнего события', async () => {
  startDemoClock();
  const demo = createGroupsDemo();
  const group = demo.getGroup(farm.item.id);
  if (!group) throw new Error('Нет деталей моковой фермы');
  let item: GroupItem = farm.item;
  server.use(
    http.get(farmPath, () => HttpResponse.json<Group>({ ...group, ...item })),
    http.get('*/api/v1/groups', ({ request }) =>
      HttpResponse.json(
        groupsPage(
          demo.getItems().map((row) => (row.id === item.id ? item : row)),
          request
        )
      )
    )
  );

  renderScreen(`/?group=${farm.item.id}`);
  await waitForScreen(() => expect(approve()).toBeDisabled());
  fireEvent.change(comment(), { target: { value: decisionComment } });

  item = { ...item, risk_score: 76 };
  await advance(3000);
  await waitForScreen(() => expect(card().getByText('Риск: высокий (76/100)')).toBeVisible());

  item = { ...item, features: { ...item.features, total_amount: 17000000 } };
  await advance(3000);
  await waitForScreen(() => expect(card().getByText(/оборот: 170 000/)).toBeVisible());
  expect(comment()).toHaveValue(decisionComment);

  item = { ...item, status: 'declined' };
  await advance(3000);
  await waitForScreen(() => expect(card().getByText('Отклонена')).toBeVisible());
  expect(within(currentGroup()).getByText('Отклонена')).toBeVisible();
  expect(card().queryByRole('button', { name: 'Подтвердить' })).not.toBeInTheDocument();
});

it('завершает начатую загрузку следующей страницы после сохранения решения', async () => {
  let finish = () => {};
  const gate = new Promise<void>((resolve) => {
    finish = resolve;
  });
  let started = false;
  server.use(
    http.get('*/api/v1/groups', async ({ request }) => {
      if (new URL(request.url).searchParams.has('cursor')) {
        started = true;
        await gate;
      }
    })
  );

  const farmLink = await openFarm();
  fireEvent.click(screen.getByRole('button', { name: 'Показать ещё' }));
  await waitFor(() => expect(started).toBe(true));
  fireEvent.change(comment(), { target: { value: decisionComment } });
  fireEvent.click(approve());
  try {
    expect(await card().findByText('Подтверждена')).toBeVisible();
    expect(within(farmLink).getByText('Подтверждена')).toBeVisible();
  } finally {
    finish();
  }
  await waitFor(() => expect(loadedGroups().getAllByRole('listitem')).toHaveLength(29));
  expect(within(farmLink).getByText('Подтверждена')).toBeVisible();
});

it('при конфликте перечитывает статус карточки и списка с сервера', async () => {
  const demo = createGroupsDemo();
  const groupId = farm.item.id;
  const detail = 'По группе уже вынесено решение';
  let status: GroupItem['status'] = 'new';
  server.use(
    http.patch(farmSolutionPath, () => {
      status = 'declined';
      return problem(409, detail);
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
  expect(await card().findByRole('alert')).toHaveTextContent(detail);
  expect(card().queryByRole('textbox', { name: 'Комментарий' })).not.toBeInTheDocument();
});
