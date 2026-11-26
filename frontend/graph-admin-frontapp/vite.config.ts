import { fileURLToPath, URL } from 'node:url';
import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const proxyTarget =
    loadEnv(mode, process.cwd(), '').VITE_ADMIN_API_PROXY_TARGET || 'http://localhost:8081';

  return {
    base: '/admin/',
    plugins: [react()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 5174,
      proxy: {
        '/admin-api': {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
  };
});
