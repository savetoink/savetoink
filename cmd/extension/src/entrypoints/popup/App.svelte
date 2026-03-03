<script lang="ts">
    import Account from "./Account.svelte";
    import Post from "./Post.svelte";
</script>

<main class="container">
    <section>
        <Post />
    </section>
    <section>
        <Account />
    </section>
</main>

<style>
    .container {
        min-width: 20rem;
        padding: 1.5rem;
    }
</style>
