import { test, expect } from '@playwright/test'

const BASE = process.env.BASE_URL ?? 'http://localhost:5173'

test.describe('Overview', () => {
  test('loads and shows stat cards', async ({ page }) => {
    await page.goto(BASE)
    await expect(page.getByText('Cluster health')).toBeVisible()
    await expect(page.getByText('Nodes')).toBeVisible()
    await expect(page.getByText('Running tasks')).toBeVisible()
    await expect(page.getByText('Active findings')).toBeVisible()
  })

  test('sidebar navigation links are present', async ({ page }) => {
    await page.goto(BASE)
    await expect(page.getByText('SwarmLens')).toBeVisible()
    for (const label of ['Stacks', 'Services', 'Tasks', 'Nodes', 'Diagnostics']) {
      await expect(page.getByText(label).first()).toBeVisible()
    }
  })

  test('navigates to diagnostics view', async ({ page }) => {
    await page.goto(BASE)
    await page.getByText('Diagnostics').click()
    await expect(page).toHaveURL(/\/diagnostics/)
  })

  test('navigates to nodes view', async ({ page }) => {
    await page.goto(BASE)
    await page.getByText('Nodes').click()
    await expect(page).toHaveURL(/\/nodes/)
  })
})

test.describe('API health', () => {
  test('healthz returns ok', async ({ request }) => {
    const res = await request.get('http://localhost:8080/api/v1/healthz')
    expect(res.ok()).toBeTruthy()
    const body = await res.json()
    expect(body.status).toBe('ok')
  })
})
