<script lang="ts">
    import Account from "./Account.svelte";
    import Post from "./Post.svelte";
</script>

<main class="container">
    <Post />
    <Account />
</main>

<style>
    .container {
        min-width: 20rem;
        padding: 1.5rem;
    }
</style>
