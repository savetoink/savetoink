<script lang="ts">
    import Settings from "./Settings.svelte";
    import Post from "./Post.svelte";
</script>

<main class="container">
    <Post />
    <Settings />
</main>

<style>
    .container {
        min-width: 20rem;
        padding: 3rem;
    }
</style>
