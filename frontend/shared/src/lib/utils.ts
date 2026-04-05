import { Temporal } from "./temporal";

export function formatDate(iso: string): string {
  return Temporal.Instant.from(iso).toLocaleString("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

export function truncate(str: string, max: number): string {
  if (str.length <= max) return str;
  return str.slice(0, max - 3) + "...";
}

export function isApiError(e: unknown): e is { error: string } {
  return (
    typeof e === "object" &&
    e !== null &&
    "error" in e &&
    typeof (e as { error: unknown }).error === "string"
  );
}

export const DeviceDomains = [
  "@kindle.com",
  "@free.kindle.com",
  "@send.kobo.com",
  "@pbsync.com",
  "@mytolino.com",
];

// Tag constants (matching backend/lib/consts/app.go)
export const MAX_TAGS = 10;
export const MAX_TAG_LENGTH = 50;
