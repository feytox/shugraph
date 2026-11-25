import { expect, test } from '@playwright/test';
import { ensureScenarioGroups, getGroup } from '../support/api';
import { expectedGroup } from '../support/scenarios';
import { cardPanel, openGroup } from '../support/ui';

const jsonHeaders = { 'content-type': 'application/json' };

test('мусор в query и path не даёт 500', async ({ request }) => {
  expect((await request.get('/api/v1/groups?cursor=broken')).status()).toBe(422);
  expect((await request.get('/api/v1/groups/not-a-uuid')).status()).toBe(422);
  expect(
    (
      await request.get('/api/v1/groups/00000000-0000-0000-0000-000000000002')
    ).status()
  ).toBe(404);
});

test('битый JSON не даёт 500', async ({ request }) => {
  const response = await request.put('/api/v1/events/logins/security-invalid-json', {
    headers: jsonHeaders,
    data: '{"occurred_at":',
  });

  expect(response.status()).toBeGreaterThanOrEqual(400);
  expect(response.status()).toBeLessThan(500);
});

test('очень длинный event_id не роняет сервис', async ({ request }) => {
  const response = await request.put(`/api/v1/events/logins/${'x'.repeat(4096)}`, {
    data: {
      occurred_at: '2026-11-23T13:00:00Z',
      account_id: 'acc-security',
      client_id: 'cli-security',
      device_id: 'dev-security',
      address_id: '203.0.113.90',
    },
  });

  expect(response.status()).toBeGreaterThanOrEqual(400);
  expect(response.status()).toBeLessThan(500);
});

test('пустой комментарий решения отклоняется', async ({ request }) => {
  const [group] = await ensureScenarioGroups(request, 'device_farm', [
    expectedGroup('device_farm'),
  ]);

  for (const comment of ['', '   ']) {
    const response = await request.patch(`/api/v1/groups/${group.id}/solution`, {
      data: { status: 'approved', decision_comment: comment },
    });
    expect(response.status()).toBe(422);
  }

  const details = await getGroup(request, group.id);
  expect(details.status).not.toBe('declined');
});

test('XSS в комментарии решения остаётся текстом', async ({ page, request }) => {
  const [group] = await ensureScenarioGroups(request, 'farm_and_address', [
    expectedGroup('farm_and_address'),
  ]);

  const payload = '<img src=x onerror="window.__xss=1">';
  const details = await getGroup(request, group.id);

  if (details.status !== 'new') {
    test.skip(true, 'на stage уже есть решение по farm_and_address; нужен свежий stage');
  }

  const decision = { comment: payload, decided_at: new Date().toISOString() };
  const resolved = { ...details, status: 'approved', is_frozen: true, decision };
  let approved = false;

  await page.route(`**/api/v1/groups/${group.id}`, (route) =>
    route.fulfill({ json: approved ? resolved : details })
  );
  await page.route('**/api/v1/groups/*/solution', (route) => {
    approved = true;
    return route.fulfill({ json: { status: 'approved', decision } });
  });

  await openGroup(page, /07\.11\.26 18:35/);

  const card = cardPanel(page);
  const comment = card.getByRole('textbox', { name: 'Комментарий' });
  await expect(comment).toBeVisible();
  await comment.fill(payload);
  await card.getByRole('button', { name: 'Подтвердить' }).click();

  await expect(card.getByText(payload, { exact: true })).toBeVisible();
  await expect(card.locator('img')).toHaveCount(0);
  expect(await page.evaluate(() => (window as unknown as { __xss?: number }).__xss)).toBeUndefined();

  const unchanged = await getGroup(request, group.id);
  expect(unchanged.status).toBe('new');
});
