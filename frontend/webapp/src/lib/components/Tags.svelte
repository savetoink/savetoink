<script lang="ts">
	import { resolve } from '$app/paths';
	import { enhance } from '$app/forms';
	import type { Article } from '@savetoink/shared';

	let {
		tags,
		articleId,
		editable = false
	}: {
		tags: Article['tags'];
		articleId?: string;
		editable?: boolean;
	} = $props();

	const canRemoveTags = $derived(editable && articleId);
</script>

{#if tags && tags.length > 0}
	<ul>
		{#each tags as tag (tag)}
			<li>
				{#if canRemoveTags}
					<form
						method="POST"
						action="?/removeTag"
						use:enhance={({ formData }) => {
							formData.set('tag', tag);
							formData.set('articleId', articleId!);
							return async ({ update }) => {
								await update();
							};
						}}
					>
						<button type="submit" aria-label={'Remove tag: ' + tag}>
							<a href={resolve(`/articles?tag=${tag}`)} onclick={(e) => e.stopPropagation()}>
								<ins>#{tag}</ins>
							</a>
							<span aria-hidden="true">&times;</span>
						</button>
					</form>
				{:else}
					<a href={resolve(`/articles?tag=${tag}`)}>
						<ins>#{tag}</ins>
					</a>
				{/if}
			</li>
		{/each}
	</ul>
{/if}

<style>
	ul {
		padding: 0;
		display: flex;
		gap: 0.8rem;
	}

	li {
		list-style: none;
	}

	form {
		margin: 0;
		padding: 0;
	}

	button {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		font: inherit;
	}

	a,
	button :global(ins) {
		text-decoration: none;
	}

	button span[aria-hidden='true'] {
		color: var(--pico-muted-color);
		font-weight: bold;
	}

	button:hover span {
		color: var(--pico-del-color);
	}
</style>
