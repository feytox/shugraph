import { expect, test, type Page } from '@playwright/test';
import { ensureScenarioGroups } from '../support/api';
import { expectedGroup } from '../support/scenarios';
import {
  cardPanel,
  expectNoHorizontalOverflow,
  graphCanvas,
  groupsPanel,
  openGroup,
  setCssViewport,
  waitForFonts,
  waitForGraph,
} from '../support/ui';

const viewports = [
  { width: 1440, height: 900 },
  { width: 1024, height: 768 },
];

async function openFarmAndAddress(page: Page) {
  await openGroup(page, /07\.11\.26 18:35/);
  await waitForFonts(page);
  await expect(graphCanvas(page)).toBeVisible();
  await waitForGraph(page);
}

test.beforeAll(async ({ request }) => {
  await ensureScenarioGroups(request, 'farm_and_address', [expectedGroup('farm_and_address')]);
});

for (const viewport of viewports) {
  test(`список, карточка и граф ${viewport.width}×${viewport.height}`, async ({ page }) => {
    await setCssViewport(page, viewport.width, viewport.height);
    await openFarmAndAddress(page);
    await expectNoHorizontalOverflow(page);

    const row = page.getByRole('link', { name: /07\.11\.26 18:35/ });
    await expect(row).toHaveScreenshot(`list-row-${viewport.width}.png`);
    await expect(cardPanel(page)).toHaveScreenshot(`card-${viewport.width}.png`);
    await expect(page.locator('main')).toHaveScreenshot(`graph-${viewport.width}.png`);
  });
}

test('свёрнутые боковые панели', async ({ page }) => {
  await openFarmAndAddress(page);

  await page.getByRole('button', { name: 'Скрыть: Подозрительные группы' }).click();
  await page.getByRole('button', { name: 'Скрыть: Карточка расследования' }).click();

  await expect(
    page.getByRole('button', { name: 'Показать: Подозрительные группы' })
  ).toBeVisible();
  await expect(page.getByRole('button', { name: 'Показать: Карточка расследования' })).toBeVisible();
  await expectNoHorizontalOverflow(page);
  await expect(page).toHaveScreenshot('collapsed-panels.png');
});

test('пустой список', async ({ page }) => {
  await page.route('**/api/v1/groups*', (route) =>
    route.fulfill({ json: { items: [], next_cursor: null } })
  );

  await page.goto('/');
  await expect(groupsPanel(page).getByText(/Подозрительных групп пока нет/)).toBeVisible();
  await expect(groupsPanel(page)).toHaveScreenshot('empty-list.png');
});

test('ошибка загрузки списка', async ({ page }) => {
  await page.route('**/api/v1/groups*', (route) => route.fulfill({ status: 500, body: '' }));

  await page.goto('/');
  await expect(groupsPanel(page).getByRole('alert')).toBeVisible();
  await expect(groupsPanel(page)).toHaveScreenshot('list-error.png');
});
