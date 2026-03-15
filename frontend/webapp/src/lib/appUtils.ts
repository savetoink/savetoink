declare const __APP_VERSION__: string;

export const APP_VERSION = __APP_VERSION__;

export function getVersionText(): string {
	return APP_VERSION;
}

export function getAppTitle(isDev: boolean): string {
	const versionTxt = getVersionText();
	return isDev ? `Save to Ink - ${versionTxt}` : 'Save to Ink';
}

export const getIsDev = (isViteDev?: boolean, isDevWorker?: string): boolean => {
	return isViteDev === true || isDevWorker === 'true';
};
