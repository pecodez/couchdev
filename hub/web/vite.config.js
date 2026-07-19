import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: { outDir: 'dist' },
  server: { proxy: { '/api': 'http://localhost:8080' } },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.js'],
    globals: true,
    css: false,
    exclude: ['**/node_modules/**', 'tests/vrt/**'],
    server: {
      deps: {
        // Process Vuetify through Vite so its CSS imports are handled.
        inline: ['vuetify'],
      },
    },
  },
})
