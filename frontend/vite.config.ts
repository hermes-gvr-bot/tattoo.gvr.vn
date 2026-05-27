import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5201,
    allowedHosts: [
      'localhost',
      '127.0.0.1',
      '.tail3840e.ts.net',
      'tmds-server-local.tail3840e.ts.net',
      'TMDSAG',
    ],
    proxy: {
      '/api': 'http://localhost:5200',
    },
  },
})
