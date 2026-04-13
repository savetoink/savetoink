const APP_URL = import.meta.env.PUBLIC_APP_URL;

// browser.action is MV3 (Chrome), browser.browserAction is MV2 (Firefox)
const browserAction = browser.action ?? browser.browserAction;

export default defineBackground(() => {
	browser.runtime.onInstalled.addListener(() => {
		browser.contextMenus.create({
			id: 'send-to-ink',
			title: 'Save to Ink',
			contexts: ['link']
		});
	});

	browser.contextMenus.onClicked.addListener((info) => {
		if (info.menuItemId !== 'send-to-ink') {
			return;
		}

		const url = info.linkUrl;
		if (!url) {
			return;
		}

		browser.tabs.create({ url: `${APP_URL}/new?url=${encodeURIComponent(url)}` });
	});

	browserAction.onClicked.addListener(async (tab) => {
		const url = tab?.url;
		if (!url || (!url.startsWith('http://') && !url.startsWith('https://'))) {
			browser.tabs.create({ url: `${APP_URL}/new` });
			return;
		}

		browser.tabs.create({ url: `${APP_URL}/new?url=${encodeURIComponent(url)}` });
	});
});
