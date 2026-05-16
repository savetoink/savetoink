import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "ink.saveto.app",
  appName: "Save to Ink",
  webDir: "dist",
  loggingBehavior: "debug",
  experimental: {
    ios: {
      spm: {
        swiftToolsVersion: "6.2",
      },
    },
  },
};

export default config;
