import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  base: './', // Using relative base for multi-path support
  plugins: [
    react(),
    {
      name: 'portal-rewrite',
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          if (req.url === '/portal' || req.url?.startsWith('/portal/')) {
            if (!req.url.includes('/assets/') && !req.url.includes('.')) {
              req.url = '/portal.html';
            }
          }
          next();
        });
      },
    }
  ],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:1816',
        changeOrigin: true,
        timeout: 300000,
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    emptyOutDir: true,
    sourcemap: false,
    rollupOptions: {
      input: {
        admin: 'index.html',
        portal: 'portal.html',
      },
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom'],
          'react-admin': ['react-admin', 'ra-data-simple-rest'],
          'echarts': ['echarts', 'echarts-for-react']
        }
      }
    }
  }
})
