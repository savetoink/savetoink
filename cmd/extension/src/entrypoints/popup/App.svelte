<script lang="ts">
  import { onMount } from 'svelte';
  import { getAuthState } from '../../lib/auth';

  let isAuthenticated = false;
  let user: any = null;
  let lastStatus: string = '';
  let sending = false;

  onMount(async () => {
    await checkAuthState();
    browser.runtime.onMessage.addListener((message) => {
      if (message.type === 'AUTH_STATE_CHANGED') {
        isAuthenticated = message.state.isAuthenticated;
        user = message.state.user;
      }
    });
  });

  async function checkAuthState() {
    const state = await getAuthState();
    isAuthenticated = state.isAuthenticated;
    user = state.user;
  }

  async function handleLogin() {
    browser.runtime.sendMessage({ type: 'LOGIN' });
  }

  async function handleLogout() {
    browser.runtime.sendMessage({ type: 'LOGOUT' });
    await checkAuthState();
  }

  async function sendCurrentPage() {
    sending = true;
    try {
      const [tab] = await browser.tabs.query({ active: true, currentWindow: true });
      if (tab?.url) {
        browser.runtime.sendMessage({ type: 'SEND_ARTICLE', url: tab.url });
        lastStatus = 'Sending article...';
      }
    } finally {
      sending = false;
    }
  }
</script>

<main>
  {#if isAuthenticated}
    <div class="user-section">
      <p>Signed in as {user?.email}</p>
      <button on:click={sendCurrentPage} disabled={sending}>
        {sending ? 'Sending...' : 'Send this page'}
      </button>
      <button on:click={handleLogout}>Logout</button>
    </div>
  {:else}
    <div class="auth-section">
      <button on:click={handleLogin}>Login with Auth0</button>
    </div>
  {/if}

  {#if lastStatus}
    <p class="status">{lastStatus}</p>
  {/if}
</main>

<style>
  main {
    padding: 1rem;
    min-width: 300px;
  }

  .user-section,
  .auth-section {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  p {
    margin: 0;
    color: #666;
  }

  button {
    padding: 0.5rem 1rem;
    font-size: 0.9rem;
  }

  .status {
    margin-top: 1rem;
    font-size: 0.85rem;
  }
</style>
