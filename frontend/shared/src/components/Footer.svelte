<script lang="ts">
	import Landing from './Landing.svelte';
	import { getVersionText } from '../lib/utils';
</script>

<footer class="container">
	<hr />
	<small>
		<Landing />
		- {getVersionText(true)} - Source on
		<a
			href="https://github.com/savetoink/savetoink"
			target="_blank"
			rel="external noopener noreferrer">GitHub</a
		>
	</small>
</footer>

<style>
	footer {
		text-align: center;
		margin-bottom: 1rem;
	}
</style>
