import { createArticle } from '../lib/api';
import { getAPIKey, getUserProfile } from '../lib/storage';

export default defineBackground(() => {
	async function showNotification(title: string, message: string) {
		try {
			await browser.notifications.create({
				type: 'basic',
				iconUrl: browser.runtime.getURL('/icon/48.png'),
				title,
				message
			});
		} catch (error) {
			console.warn('failed to show notification:', error);
		}
	}

	browser.runtime.onInstalled.addListener(() => {
		browser.contextMenus.create({
			id: 'send-to-ink',
			title: 'Save to Ink',
			contexts: ['link']
		});
	});

	browser.contextMenus.onClicked.addListener(async (info) => {
		if (info.menuItemId !== 'send-to-ink') {
			return;
		}

		const url = info.linkUrl;
		if (!url) {
			await showNotification('Error', 'No link URL found');
			return;
		}

		try {
			const apiKey = await getAPIKey();
			if (!apiKey) {
				await showNotification(
					'Authentication Required',
					'Please login to your Save to Ink account'
				);
				return;
			}

			const profile = await getUserProfile();

			await createArticle(url, profile?.auto_send || false, apiKey);
			if (profile?.auto_send) {
				await showNotification('Success', 'Article saved and sent to device');
			} else {
				await showNotification('Success', 'Article saved to your reading list');
			}
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'failed to save article';
			await showNotification('Error', errorMessage);
		}
	});
});
