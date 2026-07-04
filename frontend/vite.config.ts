import { readFileSync } from 'node:fs';
import path from 'path';
import { defineConfig } from 'vite';
import pugPlugin from 'vite-plugin-pug';

// Single source of truth for the app version: the repo-root VERSION file.
// The Go side embeds the same file (see version_embed.go at the repo root).
const appVersion = readFileSync(path.resolve(__dirname, '..', 'VERSION'), 'utf-8').trim();

export default defineConfig({
  root: '.',
  resolve: {
    alias: {
      '@wailsapp/runtime': path.resolve(__dirname, 'wailsjs/wailsjs/runtime'),
    },
  },
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  plugins: [pugPlugin()],
});
