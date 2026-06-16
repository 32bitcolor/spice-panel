import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    // Pin to a timezone west of UTC so dowLabel's missing timeZone:'UTC' is
    // caught deterministically on all CI runners (including UTC ones).
    env: { TZ: 'America/New_York' },
  },
})
