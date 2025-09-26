import { defineConfig } from 'vite';
import pugPlugin from 'vite-plugin-pug';

export default defineConfig({
  root: '.',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  plugins: [pugPlugin()],
});
