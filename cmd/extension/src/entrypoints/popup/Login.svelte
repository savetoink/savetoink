<script lang="ts">
    import SharedKeyAuth from "./SharedKeyAuth.svelte";
    import { SharedKeyBackend, Auth0Backend } from "../../lib/storage";
    import type { AuthBackendType, UserProfile } from "../../lib/storage";

    export let apiKey = "";
    export let authBackend: AuthBackendType | null = null;
    export let userProfile: UserProfile | null = null;
    export let onAuthBackendChange: (event: Event) => void | Promise<void>;
    export let onApiKeySave: (detail: {
        apiKey: string;
        profile?: UserProfile;
    }) => void | Promise<void>;
    export let onApiKeyLogout: () => void | Promise<void>;

    $: isLoggedIn = apiKey && userProfile;
</script>

{#if !isLoggedIn}
    <fieldset>
        <label>
            <input
                type="radio"
                name="auth-backend"
                value="shared_api_key"
                bind:group={authBackend}
                on:change={onAuthBackendChange}
            />
            Shared API Key
        </label>

        <label>
            <input
                type="radio"
                name="auth-backend"
                value="auth0"
                bind:group={authBackend}
                disabled
            />
            Auth0 (coming soon)
        </label>
    </fieldset>
{/if}

{#if authBackend === SharedKeyBackend}
    <SharedKeyAuth
        {apiKey}
        profile={userProfile}
        onSave={onApiKeySave}
        onLogout={onApiKeyLogout}
    />
{:else if authBackend === Auth0Backend}
    <small>Auth0 authentication is not yet implemented.</small>
{/if}
