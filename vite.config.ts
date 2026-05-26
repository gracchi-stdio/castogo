import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [tailwindcss()],
  server: {
    // Use absolute URLs to Vite dev server for CSS url() references.
    // Without this, fonts referenced in CSS via @font-face resolve
    // against the page origin (Go :8080) instead of Vite (:3000).
    origin: "http://localhost:3000",
  },
  publicDir: false,
  build: {
    manifest: true,
    emptyOutDir: false,
    outDir: "public",
    rollupOptions: {
      input: "assets/js/app.js",
      output: {
        entryFileNames: "assets/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash][extname]",
      },
    },
  },
});
