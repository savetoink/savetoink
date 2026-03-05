import { format } from "date-fns";

export function formatDate(iso: string): string {
  return format(new Date(iso), "d MMM y");
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

export const getIsDev = (isViteDev?: boolean, isDevWorker?: string): boolean => {
  return isViteDev === true || isDevWorker === 'true';
};
