<script lang="ts">
	import type { PageData } from './$types';
	import type { UserProfile } from '@savetoink/shared';

	let { data }: { data: PageData } = $props();
	let user = $derived(data?.user as UserProfile | undefined);
	let sendToDevice = $derived(user?.auto_send);
</script>

<section>
	<hgroup>
		<h1>Add Article</h1>
		<p>Save a new article to your reading list</p>
	</hgroup>

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
				/>
				<small>Enter the full URL of the article you want to save</small>
			</label>
			{#if data.user?.device_email != ''}
				<label>
					<input type="checkbox" name="sendToDevice" bind:checked={sendToDevice} />
					Send to device: <code>{user?.device_email}</code>
				</label>
			{/if}
		</fieldset>
		<button type="submit">Add</button>
	</form>
</section>
