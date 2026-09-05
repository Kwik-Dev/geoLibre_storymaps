import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

// Vitest config (separate from the single-file build config so the test
// environment doesn't force vite-plugin-singlefile onto the test bundle).
export default defineConfig({
    plugins: [react()],
    test: {
        environment: 'jsdom',
        globals: false,
        include: ['src/**/*.test.{js,jsx}'],
    },
});
