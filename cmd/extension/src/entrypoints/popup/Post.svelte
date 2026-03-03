<script lang="ts">
    import { createArticle } from "../../lib/api";
    import { getAPIKey } from "../../lib/storage";

    let url = "";
    let loading = false;
    let status: "" | "success" | "error" = "";
    let errorMessage = "";
    let statusTimeout: ReturnType<typeof setTimeout>;

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

            await createArticle(url, apiKey);
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

<form on:submit|preventDefault={handleSubmit}>
    <input
        type="url"
        name="url"
        placeholder="https://example.com/article"
        bind:value={url}
        required
    />
    <button type="submit" disabled={loading}>
        {loading ? "Adding..." : "Add"}
    </button>

    {#if status === "error"}
        <p class="error">{errorMessage}</p>
    {:else if status === "success"}
        <p class="success">article saved successfully</p>
    {/if}
</form>

<style>
    .error {
        color: #d32f2f;
    }

    .success {
        color: #388e3c;
    }
</style>
