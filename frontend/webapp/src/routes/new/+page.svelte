<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import KeyboardNav from '$lib/components/KeyboardNav.svelte';
	import { BASE_BINDINGS } from '@savetoink/shared';

	const keyboardCallbacks = {
		h: () => goto(resolve('/')),
		a: () => goto(resolve('/account'))
	};
</script>

<section>
	<hgroup>
		<h1>Add Article</h1>
		<p>Save a new article to your reading list</p>
	</hgroup>

	<!-- svelte-ignore a11y_autofocus -->
	<form method="POST" action="?/new">
		<fieldset>
			<label>
				URL
				<input
					type="url"
					name="url"
					required
					placeholder="https://example.com/article"
					autocomplete="url"
					autofocus
				/>
				<small>Enter the full URL of the article you want to save</small>
			</label>
		</fieldset>
		<button type="submit">Add</button>
	</form>
</section>

<KeyboardNav bindings={BASE_BINDINGS} callbacks={keyboardCallbacks} enabledKeys={['h', 'a']} />
