import { CapacitorShareTarget } from "@capgo/capacitor-share-target";
import { Browser } from "@capacitor/browser";
import { Capacitor } from "@capacitor/core";

const APP_URL = import.meta.env.PUBLIC_APP_URL;

if (Capacitor.isNativePlatform()) {
	CapacitorShareTarget.addListener("shareReceived", async (event) => {
		const url = event.texts?.[0];
		if (!url) {
			return;
		}

		// Only handle actual URLs
		if (!url.startsWith("http://") && !url.startsWith("https://")) {
			return;
		}

		// Open the default browser with the web app's /new page
		await Browser.open({
			url: `${APP_URL}/new?url=${encodeURIComponent(url)}`,
		});
	});
}
