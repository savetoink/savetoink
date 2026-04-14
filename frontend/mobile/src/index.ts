import { CapacitorShareTarget } from "@capgo/capacitor-share-target";
import { Capacitor } from "@capacitor/core";

const APP_URL = import.meta.env.PUBLIC_APP_URL;

if (Capacitor.isNativePlatform()) {
	CapacitorShareTarget.addListener("shareReceived", (event) => {
		const url = event.texts?.[0];
		if (!url) {
			return;
		}

		// Only handle actual URLs
		if (!url.startsWith("http://") && !url.startsWith("https://")) {
			return;
		}

		// Navigate the Capacitor webview to the web app's /new page
		window.location.href = `${APP_URL}/new?url=${encodeURIComponent(url)}`;
	});
}
