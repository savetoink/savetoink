/// <reference path="./css.d.ts" />

export async function getCSS(isDev: boolean): Promise<void> {
	if (isDev) {
		await import('@picocss/pico/css/pico.yellow.min.css');
	} else {
		await import('@picocss/pico/css/pico.min.css');
	}
}
