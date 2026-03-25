<script lang="ts">
	import TagsInput from './TagsInput.svelte';

	interface Props {
		initialTags: string[];
		onSave: (tags: string[]) => Promise<void>;
		maxTags?: number;
		maxTagLength?: number;
	}

	let { initialTags, onSave, maxTags = 10, maxTagLength = 50 }: Props = $props();

	let isEditing = $state(false);
	let editingTags = $state<string[]>([]);
	let isSaving = $state(false);
	let error = $state<string | null>(null);

	let tags = $derived(isEditing ? editingTags : initialTags);

	// Initialize editingTags when entering edit mode
	$effect(() => {
		if (isEditing && editingTags.length === 0 && initialTags.length > 0) {
			editingTags = [...initialTags];
		}
	});

	async function handleSave() {
		isSaving = true;
		error = null;

		try {
			await onSave(editingTags);
			isEditing = false;
			editingTags = [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save tags';
		} finally {
			isSaving = false;
		}
	}

	function handleCancel() {
		editingTags = [];
		isEditing = false;
		error = null;
	}

	function handleTagsChange(newTags: string[]) {
		editingTags = newTags;
		error = null;
	}
</script>

<div class="tags-editor">
	{#if isEditing}
		<div class="tags-editor-form">
			<TagsInput
				value={editingTags}
				onChange={handleTagsChange}
				disabled={isSaving}
				{maxTags}
				{maxTagLength}
				{error}
			/>
			<div class="actions">
				<button
					type="button"
					class="primary"
					onclick={handleSave}
					disabled={isSaving || editingTags.length === 0}
				>
					{isSaving ? 'Saving...' : 'Save'}
				</button>
				<button type="button" onclick={handleCancel} disabled={isSaving}> Cancel </button>
			</div>
		</div>
	{:else}
		<div
			class="tags-display"
			onclick={() => (isEditing = true)}
			role="button"
			tabindex="0"
			onkeydown={(e) => {
				if (e.key === 'Enter' || e.key === ' ') {
					e.preventDefault();
					isEditing = true;
				}
			}}
		>
			{#if tags.length > 0}
				<ul>
					{#each tags as tag (tag)}
						<li><ins>#{tag}</ins></li>
					{/each}
				</ul>
			{:else}
				<span class="no-tags">Add tags...</span>
			{/if}
			<button
				type="button"
				class="edit-button"
				aria-label="Edit tags"
				onclick={(e) => {
					e.stopPropagation();
					isEditing = true;
				}}
			>
				✎
			</button>
		</div>
	{/if}
</div>

<style>
	.tags-editor {
		width: 100%;
	}

	.tags-editor-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.actions {
		display: flex;
		gap: 0.5rem;
	}

	.tags-display {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem;
		border-radius: var(--pico-border-radius);
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.tags-display:hover {
		background-color: var(--pico-card-background-color);
	}

	.tags-display:focus-visible {
		outline: 2px solid var(--pico-primary-focus);
		outline-offset: 2px;
	}

	.tags-display ul {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		padding: 0;
		margin: 0;
		list-style: none;
		flex: 1;
	}

	.tags-display li {
		display: inline-block;
		background-color: var(--pico-primary-background);
		color: var(--pico-color);
		border-radius: var(--pico-border-radius);
		padding: 0.25rem 0.5rem;
	}

	.tags-display ins {
		text-decoration: none;
	}

	.no-tags {
		color: var(--muted-color);
		font-style: italic;
	}

	.edit-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		padding: 0;
		border: none;
		background: transparent;
		color: var(--muted-color);
		cursor: pointer;
		font-size: 1rem;
		border-radius: 50%;
		transition:
			background-color 0.2s,
			color 0.2s;
		flex-shrink: 0;
	}

	.edit-button:hover {
		background-color: var(--pico-primary-background);
		color: var(--pico-color);
	}
</style>
