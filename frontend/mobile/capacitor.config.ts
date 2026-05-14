import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "ink.saveto.app",
  appName: "Save to Ink",
  webDir: "dist",
  loggingBehavior: "debug",
  plugins: {
    CapacitorShareTarget: {
      appGroupId: "group.ink.saveto.app",
    },
  },
  experimental: {
    ios: {
      spm: {
        swiftToolsVersion: "6.2",
      },
    },
  },
};

export default config;
