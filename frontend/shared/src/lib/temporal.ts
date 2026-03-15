import { Temporal as PolyfillTemporal } from "temporal-polyfill";

export const Temporal =
  (globalThis as typeof globalThis & { Temporal?: typeof PolyfillTemporal }).Temporal ?? PolyfillTemporal;
