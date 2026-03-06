import { format } from "date-fns";

declare const __APP_VERSION__: string;
declare const __BUILD_DATE__: string;
declare const __GIT_HASH__: string;

export const APP_VERSION = __APP_VERSION__;
export const BUILD_DATE = __BUILD_DATE__;
export const GIT_HASH = __GIT_HASH__;

export function getVersionText(isDev: boolean): string {
	if (isDev) {
		return `${APP_VERSION}-${BUILD_DATE}-${GIT_HASH}`;
	}
	return APP_VERSION;
}

export function getAppTitle(isDev: boolean): string {
	const versionTxt = getVersionText(isDev);
	return isDev ? `Save to Ink - ${versionTxt}` : 'Save to Ink';
}

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
