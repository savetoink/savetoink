<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	let sendToDevice = $derived(data?.user?.autoSend || false);
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
			<!-- <label>
				Tags
				<input
					type="text"
					name="tags"
					maxlength="200"
					placeholder="tech, reading, tutorial"
					autocomplete="off"
				/>
				<small>Optional comma-separated tags (e.g., tech, reading, tutorial)</small>
			</label> -->
			{#if data?.user?.deviceEmail}
				<label>
					<input type="checkbox" name="sendToDevice" bind:checked={sendToDevice} />
					Send to device (<code>{data.user.deviceEmail}</code>)
				</label>
			{/if}
		</fieldset>
		<button type="submit">Add</button>
	</form>
</section>
