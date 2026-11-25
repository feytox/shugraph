import { expect, test } from '@playwright/test';
import { ensureScenarioGroups, getGroup } from '../support/api';
import { expectedGroup } from '../support/scenarios';
import {
  cardPanel,
  graphCanvas,
  openGroup,
  waitForFonts,
  waitForGraph,
} from '../support/ui';

test('device_farm: путь от списка до решения аналитика', async ({ page, request }) => {
  const [group] = await ensureScenarioGroups(request, 'device_farm', [
    expectedGroup('device_farm'),
  ]);

  await openGroup(page, /05\.11\.26 10:29/);
  await waitForFonts(page);

  const row = page.getByRole('link', { name: /05\.11\.26 10:29/ });
  await expect(row).toContainText('Ферма');
  await expect(row.getByLabel('Риск высокий: 75')).toBeVisible();
  await expect(row).toContainText(/4\s+счёта/);
  await expect(row).toContainText(/161\s*000/);

  const card = cardPanel(page);
  await expect(card.getByText('Ферма', { exact: true })).toBeVisible();
  await expect(card.getByLabel('Риск высокий: 75')).toBeVisible();
  await expect(card).toContainText('счетов');
  await expect(card).toContainText('клиентов');
  await expect(card).toContainText('28 мин');
  await expect(card).toContainText(/161\s*000/);
  await expect(card.getByText(/разные клиенты \(всего 3\)/)).toBeVisible();
  await expect(card.getByText(/05\.11\.26, 10:01–10:29 UTC/)).toBeVisible();

  const risk = card.getByRole('table', { name: 'Из чего риск' });
  await expect(risk).toContainText('Схема «ферма»');
  await expect(risk).toContainText('+50');
  await expect(risk).toContainText('+15');
  await expect(risk).toContainText('+0');
  await expect(risk).toContainText('+10');

  await expect(graphCanvas(page)).toBeVisible();
  await waitForGraph(page);
  await expect(page.getByText('перевод', { exact: true })).toBeVisible();
  await expect(page.getByText('вход', { exact: true })).toBeVisible();
  await expect(page.getByText('время в UTC')).toBeVisible();

  const commentText = 'E2E: подтверждено автоматическим тестом';
  const comment = card.getByRole('textbox', { name: 'Комментарий' });
  const approve = card.getByRole('button', { name: 'Подтвердить' });

  if (await comment.isVisible()) {
    await expect(approve).toBeDisabled();
    await comment.fill(commentText);
    await expect(approve).toBeEnabled();
    await approve.click();

    await expect(card.getByText('Подтверждена')).toBeVisible();
    await expect(card.getByText(commentText)).toBeVisible();
    await expect(row.getByText('Подтверждена')).toBeVisible();

    const details = await getGroup(request, group.id);
    expect(details.status).toBe('approved');
    expect(details.is_frozen).toBe(true);
    expect(details.decision?.comment).toBe(commentText);
  } else {
    test.info().annotations.push({
      type: 'note',
      description:
        'Группа уже решена на этом stage; для проверки клика по решению нужен свежий stage.',
    });
    await expect(card.getByText('Подтверждена')).toBeVisible();
  }
});
