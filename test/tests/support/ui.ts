import { expect, type Page } from '@playwright/test';

export const groupsPanel = (page: Page) =>
  page.getByRole('complementary', { name: 'Подозрительные группы' });

export const cardPanel = (page: Page) =>
  page.getByRole('complementary', { name: 'Карточка расследования' });

export const graphCanvas = (page: Page) => page.getByRole('img', { name: 'Граф группы' });

export async function setCssViewport(page: Page, width: number, height: number): Promise<void> {
  await page.setViewportSize({ width, height });
  const dpr = await page.evaluate(() => window.devicePixelRatio);
  if (Math.abs(dpr - 1) > 0.001) {
    await page.setViewportSize({
      width: Math.round(width * dpr),
      height: Math.round(height * dpr),
    });
  }
}

export async function openGroup(page: Page, rowText: RegExp): Promise<void> {
  await page.goto('/');
  await expect(groupsPanel(page).getByRole('list', { name: 'Список групп' })).toBeVisible();

  const row = page.getByRole('link', { name: rowText });
  await expect(row).toBeVisible();
  await row.click();
  await expect(cardPanel(page)).toBeVisible();
}

export async function waitForFonts(page: Page): Promise<void> {
  await page.evaluate(() => document.fonts.ready);
}

export async function waitForGraph(page: Page): Promise<void> {
  await expect(graphCanvas(page)).toBeVisible();
  await page.waitForFunction(() => {
    const canvas = document.querySelector('canvas');
    return Boolean(canvas && canvas.width > 0 && canvas.height > 0);
  });
  await page.waitForTimeout(300);
}

export async function expectNoHorizontalOverflow(page: Page): Promise<void> {
  const overflow = await page.locator('#root > div').evaluate((element) => {
    return element.scrollWidth - element.clientWidth;
  });
  expect(overflow).toBeLessThanOrEqual(1);
}
