<script lang="ts">
    import SharedKeyAuth from "./SharedKeyAuth.svelte";
    import Auth0Auth from "./Auth0Auth.svelte";
    import { SharedKeyBackend, Auth0Backend } from "../../lib/storage";
    import type { AuthBackendType, UserProfile } from "../../lib/storage";

    export let apiKey = "";
    export let authBackend: AuthBackendType | null = null;
    export let userProfile: UserProfile | null = null;
    export let onApiKeySave: (detail: {
        apiKey: string;
        profile?: UserProfile;
    }) => void | Promise<void>;
    export let onApiKeyLogout: () => void | Promise<void>;
</script>

{#if authBackend === SharedKeyBackend}
    <SharedKeyAuth
        {apiKey}
        profile={userProfile}
        onSave={onApiKeySave}
        onLogout={onApiKeyLogout}
    />
{:else if authBackend === Auth0Backend}
    <Auth0Auth
        {apiKey}
        profile={userProfile}
        onSave={onApiKeySave}
        onLogout={onApiKeyLogout}
    />
{/if}
