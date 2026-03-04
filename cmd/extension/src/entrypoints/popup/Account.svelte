<script lang="ts">
    import { onMount } from "svelte";
    import Login from "./Login.svelte";
    import {
        getAPIKey,
        saveAPIKey,
        clearAPIKey,
        getUserProfile,
        saveUserProfile,
        clearUserProfile,
        SharedKeyBackend,
        Auth0Backend,
    } from "../../lib/storage";
    import { API_URL } from "../../lib/api";
    import type { AuthBackendType, UserProfile } from "../../lib/storage";

    let apiKey = "";
    let authBackend: AuthBackendType =
        (import.meta.env.PUBLIC_AUTH_BACKEND as AuthBackendType) ||
        SharedKeyBackend;
    let userProfile: UserProfile | null = null;

    onMount(async () => {
        const [savedKey, savedProfile] = await Promise.all([
            getAPIKey(),
            getUserProfile(),
        ]);
        if (savedKey) {
            apiKey = savedKey;
        }
        userProfile = savedProfile;
    });

    async function handleApiKeySave(detail: {
        apiKey: string;
        profile?: UserProfile;
    }) {
        await saveAPIKey(detail.apiKey);
        apiKey = detail.apiKey;
        if (detail.profile) {
            await saveUserProfile(detail.profile);
            userProfile = detail.profile;
        }
    }

    async function handleApiKeyLogout() {
        await Promise.all([clearAPIKey(), clearUserProfile()]);
        apiKey = "";
        userProfile = null;
    }
</script>

<h1>Account</h1>

<section>
    <Login
        {apiKey}
        {authBackend}
        {userProfile}
        onApiKeySave={handleApiKeySave}
        onApiKeyLogout={handleApiKeyLogout}
    />
</section>

<section>
    <ul>
        {#if userProfile}
            {#if userProfile.email}
                <li>
                    <strong>Email:</strong>
                    {userProfile.email}
                </li>
            {/if}

            {#if userProfile.device_email}
                <li>
                    <strong>Device email:</strong>
                    {userProfile.device_email || "Not set"}
                </li>
            {/if}

            {#if userProfile.auto_send}
                <li>
                    <strong>Auto-send:</strong>
                    {userProfile.auto_send ? "Enabled" : "Disabled"}
                </li>
            {/if}
        {/if}
        {#if import.meta.env.DEV}
            <li>
                <strong>API URL:</strong>
                {API_URL}
            </li>
        {/if}
    </ul>
</section>

<style>
    ul li {
        list-style: none;
    }
</style>
