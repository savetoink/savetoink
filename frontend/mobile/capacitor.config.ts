import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "ink.saveto.app",
  appName: "Save to Ink",
  webDir: "dist",
  loggingBehavior: "debug",
  server: {
    url: process.env.PUBLIC_APP_URL,
    cleartext: false,
    allowNavigation: [
      "https://app.saveto.ink",
      "https://app.dev.saveto.ink",
      "https://auth.saveto.ink",
    ],
  },
  plugins: {
    CapacitorShareTarget: {
      appGroupId: "group.ink.saveto.app",
    },
  },
};

export default config;
