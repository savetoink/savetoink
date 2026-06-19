import { defineConfig } from "vite";

export default defineConfig({
  base: "./",
  build: {
    target: "esnext",
    outDir: "dist",
    emptyOutDir: true,
  },
});
