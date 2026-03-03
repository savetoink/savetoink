import { defineConfig } from "wxt";

// See https://wxt.dev/api/config.html
export default defineConfig({
  srcDir: "src",
  outDir: "dist",
  modules: ["@wxt-dev/module-svelte"],
  manifest: {
    permissions: ["storage", "tabs", "contextMenus", "identity"],
  },
  vite(env) {
    return {
      envPrefix: "PUBLIC_",
    };
  },
});
