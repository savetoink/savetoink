export default defineContentScript({
  matches: ['<all_urls>'],
  main() {
    browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
      if (message.type === 'SEND_ARTICLE') {
        const url = window.location.href;
        browser.runtime.sendMessage({ type: 'SEND_ARTICLE', url });
      }
    });
  },
});
