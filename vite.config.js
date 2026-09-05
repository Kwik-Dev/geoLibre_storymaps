import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import { viteSingleFile } from 'vite-plugin-singlefile';

// base './' so the built dist/index.html can be opened directly from disk
// (file://). vite-plugin-singlefile inlines the JS + CSS into that one HTML
// file — ES module <script src> fetches are blocked on file:// origins, but
// an inline module script is not. Media (audio/images/videos) stay as files
// next to index.html, referenced with relative paths from the story JSONs.
//
// When VITE_BASE_PATH is set (e.g. "/maps"), the build base becomes that
// subpath so asset URLs are emitted under it for subpath deployments.
export default defineConfig(({ mode }) => {
  // Read the backend port from .env (APP_PORT) so the dev proxy stays in sync
  // with the Go server. Override with BACKEND_URL for a full origin.
  const env = loadEnv(mode, process.cwd(), '');
  const backendUrl = env.BACKEND_URL || `http://localhost:${env.APP_PORT || '8080'}`;
  const basePath = (env.VITE_BASE_PATH || '').trim();
  const base = basePath && basePath !== '/'
    ? basePath.replace(/\/+$/, '') + '/'
    : './';

  return {
    plugins: [react(), viteSingleFile()],
    base,
    server: {
      port: 5173,
      proxy: {
        // Dev-only proxy → Go backend (port from .env APP_PORT, default 8080).
        // /api (REST) and /media (uploaded media) are forwarded with the same path.
        '/api': backendUrl,
        '/media': backendUrl,
      },
    },
  };
});
