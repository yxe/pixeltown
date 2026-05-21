import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// dev-server proxy: /api/* and /ws/* are forwarded to the Go
// server in docker. Keeps the frontend's fetch calls relative so
// the same code works against any backend host.
export default defineConfig({
  plugins: [svelte()],
  server: {
    proxy: {
      '/api': 'http://localhost:4000',
      '/ws': { target: 'ws://localhost:4000', ws: true },
    },
  },
})
