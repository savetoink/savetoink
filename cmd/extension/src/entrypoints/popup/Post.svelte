<script lang="ts">
    import { createArticle, sendArticle } from "../../lib/api";
    import { getAPIKey, getUserProfile } from "../../lib/storage";
    import type { UserProfile } from "../../lib/storage";

    let url = $state("");
    let loading = $state(false);
    let status: "" | "success" | "error" = $state("");
    let errorMessage = $state("");
    let statusTimeout: ReturnType<typeof setTimeout>;
    let profile: UserProfile | null = $state(null);
    let sendToDevice = $state(false);

    async function loadProfile() {
        const apiKey = await getAPIKey();
        if (apiKey) {
            profile = await getUserProfile();
            sendToDevice = profile?.auto_send || false;
        }
    }

    loadProfile();

    function clearStatus() {
        status = "";
        errorMessage = "";
    }

    async function handleSubmit(event: Event) {
        event.preventDefault();

        if (!url) {
            return;
        }

        clearTimeout(statusTimeout);
        loading = true;
        status = "";
        errorMessage = "";

        try {
            const apiKey = await getAPIKey();
            if (!apiKey) {
                errorMessage = "please login first";
                status = "error";
                loading = false;
                statusTimeout = setTimeout(clearStatus, 5000);
                return;
            }

            const article = await createArticle(url, apiKey);

            if (sendToDevice) {
                const sendResponse = await sendArticle(article.id, apiKey);
                if (!sendResponse.ok) {
                    throw new Error("failed to send article to device");
                }
            }

            status = "success";

            statusTimeout = setTimeout(() => {
                window.close();
            }, 2000);
        } catch (error) {
            errorMessage =
                error instanceof Error
                    ? error.message
                    : "failed to create article";
            status = "error";
            loading = false;
            statusTimeout = setTimeout(clearStatus, 5000);
        } finally {
            if (status !== "error") {
                loading = false;
            }
        }
    }
</script>

<h1>Add Article</h1>

<p>Save a new article to your reading list</p>

<form onsubmit={handleSubmit}>
    <input
        type="url"
        name="url"
        placeholder="https://example.com/article"
        bind:value={url}
        required
    />
    {#if profile?.device_email}
        <label>
            <input type="checkbox" bind:checked={sendToDevice} />
            Send to reader device
        </label>
    {/if}
    <button type="submit" disabled={loading}>
        {loading ? "Adding..." : "Add"}
    </button>

    {#if status === "error"}
        <p aria-invalid="true" class="error">
            {errorMessage}
        </p>
    {:else if status === "success"}
        {#if sendToDevice}
            <ins>article saved and delivered successfully</ins>
        {:else}
            <ins>article saved successfully</ins>
        {/if}
    {/if}
</form>
