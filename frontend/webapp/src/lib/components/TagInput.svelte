<script lang="ts">
	const MAX_TAGS = 10;
	const MAX_TAG_LENGTH = 50;

	let {
		tags = [],
		onAdd,
		onRemove
	}: {
		tags?: string[];
		onAdd?: (tags: string[]) => void;
		onRemove?: (tag: string) => void;
	} = $props();

	let tagsInput = $state('');
	let tagsError = $state<string | null>(null);

	let parsedTags = $derived(() => {
		if (!tagsInput) return [];
		return tagsInput
			.split(',')
			.map((tag) => tag.trim())
			.filter((tag) => tag.length > 0);
	});

	function validateTags(newTags: string[]): string | null {
		if (newTags.length > MAX_TAGS) {
			return `Maximum ${MAX_TAGS} tags allowed per article`;
		}
		for (const tag of newTags) {
			if (tag.length > MAX_TAG_LENGTH) {
				return `Tag "${tag}" exceeds maximum length of ${MAX_TAG_LENGTH} characters`;
			}
		}
		return null;
	}

	async function handleAddTags() {
		tagsError = null;

		// Validate tags before adding
		const newTags = parsedTags();
		const validationError = validateTags(newTags);
		if (validationError) {
			tagsError = validationError;
			return;
		}

		// Check for duplicates with existing tags
		const combinedTags = [...tags, ...newTags];
		const uniqueTags = Array.from(new Set(combinedTags));
		const duplicateError = validateTags(uniqueTags);
		if (duplicateError) {
			tagsError = duplicateError;
			return;
		}

		if (onAdd) {
			await onAdd(newTags);
			tagsInput = '';
		}
	}

	async function handleRemoveTag(tag: string) {
		if (onRemove) {
			await onRemove(tag);
		}
	}
</script>

<section class="tag-input">
	<label>
		Existing Tags
		{#if tags.length > 0}
			<ul>
				{#each tags as tag (tag)}
					<li>
						<button
							type="button"
							class="tag-remove"
							onclick={() => handleRemoveTag(tag)}
							title={`Remove tag: ${tag}`}
						>
							{tag} <span aria-hidden="true">×</span>
						</button>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="no-tags">No tags yet</p>
		{/if}
	</label>
	{#if onAdd}
		<label>
			Add Tags
			<input
				type="text"
				placeholder="reading, tech, tutorial"
				bind:value={tagsInput}
				disabled={false}
			/>
			<small>
				Comma-separated list of tags (max {MAX_TAGS} tags total, {MAX_TAG_LENGTH} characters each).
				{(parsedTags().length > 0 && ` Current input: ${parsedTags().length}`) || undefined}
			</small>
		</label>
		{#if tagsError}
			<p class="error" role="alert">{tagsError}</p>
		{/if}
		<button
			type="button"
			class="add-tags-button"
			onclick={handleAddTags}
			disabled={parsedTags().length === 0}
		>
			Add Tags
		</button>
	{/if}
</section>

<style>
	.tag-input {
		margin: 1rem 0;
	}

	ul {
		padding: 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		list-style: none;
		margin: 0.5rem 0;
	}

	li {
		list-style: none;
	}

	.tag-remove {
		background: var(--pico-primary);
		color: var(--pico-primary-inverse);
		border: none;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.875rem;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.tag-remove:hover {
		opacity: 0.9;
	}

	.tag-remove[aria-hidden='true'] {
		font-weight: bold;
	}

	.no-tags {
		color: var(--pico-muted-color);
		font-style: italic;
		margin: 0.5rem 0;
	}

	.add-tags-button {
		width: 100%;
		margin-top: 0.5rem;
	}

	.add-tags-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.error {
		color: var(--pico-del-color);
		margin: 0.5rem 0;
	}
</style>
