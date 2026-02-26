import { login, logout, getAccessToken, getAuthState } from '../lib/auth';
import { sendArticle } from '../lib/api';

export default defineBackground(() => {
  console.log('SaveToInk extension background service worker started');

  browser.contextMenus.create({
    id: 'send-article',
    title: 'Send to SaveToInk',
    contexts: ['page', 'selection', 'link'],
  });

  browser.contextMenus.onClicked.addListener(async (info, tab) => {
    if (info.menuItemId === 'send-article') {
      const url = info.linkUrl || tab?.url;
      if (url) {
        await sendCurrentPageUrl(url);
      }
    }
  });

  browser.runtime.onMessage.addListener(async (message, sender, sendResponse) => {
    switch (message.type) {
      case 'LOGIN':
        await login();
        break;

      case 'LOGOUT':
        await logout();
        break;

      case 'SEND_ARTICLE':
        if (message.url) {
          await sendCurrentPageUrl(message.url);
        }
        break;

      case 'GET_AUTH_STATUS':
        const state = await getAuthState();
        sendResponse(state);
        return true;

      case 'AUTH_CALLBACK':
        try {
          const { handleRedirectCallback } = await import('../lib/auth');
          await handleRedirectCallback();
          sendResponse({ success: true });
        } catch (error) {
          console.error('Auth callback failed:', error);
          sendResponse({ success: false, error: String(error) });
        }
        return true;

      default:
        console.warn('Unknown message type:', message.type);
    }
  });

  browser.alarms.create('tokenRefresh', { periodInMinutes: 5 });

  browser.alarms.onAlarm.addListener(async (alarm) => {
    if (alarm.name === 'tokenRefresh') {
      try {
        const { isAuthenticated, refreshTokens } = await import('../lib/auth');
        if (await isAuthenticated()) {
          await refreshTokens();
        }
      } catch (error) {
        console.error('Token refresh failed:', error);
      }
    }
  });

  async function sendCurrentPageUrl(url: string): Promise<void> {
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        showNotification('Not authenticated', 'Please log in to SaveToInk');
        return;
      }

      const response = await sendArticle(url, accessToken);
      showNotification('Article sent', response.message);
    } catch (error) {
      console.error('Failed to send article:', error);
      showNotification('Error', `Failed to send article: ${error}`);
    }
  }

  function showNotification(title: string, message: string): void {
    browser.notifications.create({
      type: 'basic',
      iconUrl: '/icon/icon-128.png',
      title,
      message,
    });
  }
});
