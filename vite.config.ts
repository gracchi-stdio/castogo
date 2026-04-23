import browserslist from 'browserslist';
import { browserslistToTargets } from 'lightningcss';
import { defineConfig } from 'vite';

const backend = 'http://localhost:8080';

export default defineConfig({
  server: {
    port: 3000,
    proxy: {
      '/shoelace': backend,
      '/admin': backend,
      '/login': backend,
      '/register': backend,
      '/logout': backend,
      '/healthcheck': backend,
    },
  },
  css: {
    transformer: 'lightningcss',
    lightningcss: {
      targets: browserslistToTargets(browserslist('>= 0.25%')),
    },
  },
  build: {
    cssMinify: 'lightningcss',
    outDir: 'public',
    emptyOutDir: false,
    rollupOptions: {
      input: 'assets/js/app.js',
      output: {
        entryFileNames: 'js/app.js',
        assetFileNames: (assetInfo) => {
          if (assetInfo.name && assetInfo.name.endsWith('.css')) {
            return 'css/app.css';
          }
          return 'assets/[name][extname]';
        },
      },
    },
  },
});
