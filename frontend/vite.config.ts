import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import http from 'http'

const resetBackendPlugin = () => {
  return {
    name: 'reset-backend-plugin',
    configureServer(server) {
      server.httpServer?.once('listening', () => {
        const req = http.request({
          hostname: 'localhost',
          port: 8080,
          path: '/api/game/create',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          }
        }, () => {
          console.log('\x1b[36m%s\x1b[0m', '>> Poker backend game automatically reset on Vite dev server startup.');
        });
        req.on('error', () => {
        });
        req.write(JSON.stringify({ small_blind: 10, big_blind: 20 }));
        req.end();
      });
    }
  }
}

export default defineConfig({
  plugins: [react(), resetBackendPlugin()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
      }
    }
  }
})
