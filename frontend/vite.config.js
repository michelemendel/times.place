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
      },
    },
  }
});

