<script lang="ts">
    import { getProfile } from "../../lib/api";
    import type { UserProfile } from "../../lib/storage";

    export let apiKey = "";
    export let profile: UserProfile | null = null;
    export let onSave: (detail: {
        apiKey: string;
        profile?: UserProfile;
    }) => void | Promise<void>;
    export let onLogout: () => void | Promise<void>;

    let localKey = apiKey;
    let saveStatus = "";

    $: localKey = apiKey;
    $: isLoggedIn = apiKey && profile;

    async function handleSave() {
        try {
            saveStatus = "validating";

            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 10000);

            const response = await getProfile(localKey);

            clearTimeout(timeoutId);

            if (response.ok) {
                await onSave({ apiKey: localKey, profile: response.profile });
                saveStatus = "saved";
                setTimeout(() => {
                    saveStatus = "";
                }, 2000);
            } else if (response.status === 401 || response.status === 403) {
                saveStatus = "invalid";
            } else {
                saveStatus = "error";
            }
        } catch (error) {
            saveStatus = "error";
            console.error("failed to validate API key:", error);
        }
    }

    async function handleLogout() {
        await onLogout();
    }
</script>

{#if !isLoggedIn}
    <form on:submit|preventDefault={handleSave}>
        <label for="api-key">API Key</label>
        <input
            id="api-key"
            type="password"
            bind:value={localKey}
            placeholder="Enter your shared API key"
            required
            minlength="1"
        />

        <button type="submit">Login</button>
        {#if saveStatus === "validating"}
            <small>validating API key...</small>
        {:else if saveStatus === "saved"}
            <small>logged in successfully</small>
        {:else if saveStatus === "invalid"}
            <small>invalid API key</small>
        {:else if saveStatus === "error"}
            <small>failed to save API key</small>
        {/if}
    </form>
{:else}
    <button type="button" on:click={handleLogout}>Logout</button>
{/if}
