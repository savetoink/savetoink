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
					<div class="tag-wrapper">
						<a href={resolve(`/articles?tag=${tag}`)} class="tag-link">
							<ins>#{tag}</ins>
						</a>
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
							<button type="submit" aria-label={'Remove ' + tag + ' tag'} class="remove-tag-btn">
								<span aria-hidden="true">&times;</span>
							</button>
						</form>
					</div>
				{:else}
					<a href={resolve(`/articles?tag=${tag}`)} class="tag-link">
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

	.tag-wrapper {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		margin: 0;
		padding: 0;
	}

	.tag-link {
		text-decoration: none;
	}

	.remove-tag-btn {
		background: none;
		border: none;
		padding: 0;
		margin: 0;
		cursor: pointer;
	}

	.remove-tag-btn span[aria-hidden='true'] {
		color: var(--pico-muted-color);
		font-weight: bold;
	}

	.remove-tag-btn:hover span {
		color: var(--pico-del-color);
	}
</style>
