import { createArticle, sendArticle } from '../lib/api';
import { getAPIKey, getUserProfile } from '../lib/storage';

export default defineBackground(() => {
	async function showToast(
		title: string,
		message: string,
		variant: 'success' | 'error' = 'success'
	) {
		try {
			const [tab] = await browser.tabs.query({ active: true, currentWindow: true });
			if (tab && tab.id) {
				await browser.tabs.sendMessage(tab.id, {
					type: 'show-toast',
					title,
					message,
					variant
				});
			}
		} catch (error) {
			console.warn('failed to show toast:', error);
		}
	}

	browser.runtime.onInstalled.addListener(() => {
		browser.contextMenus.create({
			id: 'send-to-ink',
			title: 'Send to Ink',
			contexts: ['link']
		});
	});

	browser.contextMenus.onClicked.addListener(async (info) => {
		if (info.menuItemId !== 'send-to-ink') {
			return;
		}

		const url = info.linkUrl;
		if (!url) {
			await showToast('Error', 'No link URL found', 'error');
			return;
		}

		try {
			const apiKey = await getAPIKey();
			if (!apiKey) {
				await showToast(
					'Authentication Required',
					'Please login to your Save to Ink account',
					'error'
				);
				return;
			}

			const profile = await getUserProfile();

			const article = await createArticle(url, null, apiKey);

			if (profile?.autoSend) {
				await sendArticle(article.id, apiKey);

				await showToast('Success', 'Article saved and sent to device', 'success');
			} else {
				await showToast('Success', 'Article saved to your reading list', 'success');
			}
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'failed to save article';
			await showToast('Error', errorMessage, 'error');
		}
	});
});
