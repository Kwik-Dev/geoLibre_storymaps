import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { viteSingleFile } from 'vite-plugin-singlefile';

// base './' so the built dist/index.html can be opened directly from disk
// (file://). vite-plugin-singlefile inlines the JS + CSS into that one HTML
// file — ES module <script src> fetches are blocked on file:// origins, but
// an inline module script is not. Media (audio/images/videos) stay as files
// next to index.html, referenced with relative paths from the story JSONs.
export default defineConfig({
  plugins: [react(), viteSingleFile()],
  base: './',
  server: {
    port: 5173,
  },
});
