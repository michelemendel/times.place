import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    host: '0.0.0.0', // Allow access from other devices on the network
    port: 5173, // Default Vite port (you can change this if needed)
    proxy: {
      // Proxy API requests to backend during development
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // Handle connection errors gracefully (suppress noisy errors when backend isn't running)
        configure: (proxy, _options) => {
          // @ts-ignore - http-proxy-middleware event handler types
          proxy.on('error', (/** @type {NodeJS.ErrnoException} */ err, /** @type {any} */ _req, /** @type {import('http').ServerResponse | undefined} */ res) => {
            // Suppress connection errors (expected when backend isn't running)
            // These are handled gracefully by the frontend API client
            const isConnectionError = 
              err.code === 'ECONNREFUSED' || 
              err.code === 'ECONNRESET' || 
              err.code === 'EPIPE' ||
              err.message?.includes('socket hang up');
            
            if (!isConnectionError) {
              console.error('Proxy error:', err.message);
            }
            
            // Send a proper error response if possible
            if (res && !res.headersSent) {
              res.writeHead(502, {
                'Content-Type': 'application/json',
              });
              res.end(JSON.stringify({
                error: {
                  code: 'backend_unavailable',
                  message: 'Backend server is not running. Please start it with: make brun'
                }
              }));
            }
          });
        },
      },
    },
  }
});

