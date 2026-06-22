import { test, expect } from '@playwright/test'

// Story iframe base — Storybook renders each story in isolation here.
const storyUrl = (id) =>
  `/iframe.html?id=${id}&viewMode=story&globals=backgrounds.value:%23121212`

// Wait for the dialog to be visible and for Vue's nextTick to flush
// (stories set internal state in mounted + nextTick).
async function loadStory(page, id) {
  await page.goto(storyUrl(id))
  await page.locator('.v-dialog').waitFor({ state: 'visible' })
  // Extra tick for stories that set form state after mount.
  await page.waitForTimeout(100)
}

const STORIES = [
  {
    id: 'components-newprojectdialog--clone-empty',
    name: 'clone-empty',
  },
  {
    id: 'components-newprojectdialog--clone-with-url',
    name: 'clone-with-url',
  },
  {
    id: 'components-newprojectdialog--greenfield-empty',
    name: 'greenfield-empty',
  },
  {
    id: 'components-newprojectdialog--greenfield-named',
    name: 'greenfield-named',
  },
  {
    id: 'components-newprojectdialog--invalid-name',
    name: 'invalid-name',
  },
  {
    id: 'components-newprojectdialog--loading',
    name: 'loading',
  },
  {
    id: 'components-newprojectdialog--api-error',
    name: 'api-error',
  },
]

for (const { id, name } of STORIES) {
  test(name, async ({ page }) => {
    await loadStory(page, id)
    await expect(page).toHaveScreenshot(`${name}.png`)
  })
}
