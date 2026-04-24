import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    tailwindcss(),
  ],
  publicDir: false,
  build: {
    emptyOutDir: false,
    rollupOptions: {
      input: {
        js: 'assets/js/app.js',
      },
      output: {
        dir: 'public',
        entryFileNames: 'js/app.js',
        assetFileNames: 'css/app.css',
      },
    },
  },
})
