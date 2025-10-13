import path from 'path';
import { defineConfig } from 'vite';
import pugPlugin from 'vite-plugin-pug';

export default defineConfig({
  root: '.',
  resolve: {
    alias: {
      '@wailsapp/runtime': path.resolve(__dirname, 'wailsjs/wailsjs/runtime'),
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  plugins: [pugPlugin()],
});
