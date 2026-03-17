import { CapacitorShareTarget } from "@capgo/capacitor-share-target";

const APP_URL = import.meta.env.PUBLIC_APP_URL;
if (!APP_URL) {
  document.body.innerHTML =
    "<h1>Error: PUBLIC_APP_URL not set at build time</h1>";
  console.error("PUBLIC_APP_URL is not defined at build time");
  throw new Error("PUBLIC_APP_URL must be set at build time");
}

const { version } = await CapacitorShareTarget.getPluginVersion();
console.log(`🚀 Share listener v${version} registered`);
globalThis.location.href = APP_URL;
