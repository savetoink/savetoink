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
};

export default config;
