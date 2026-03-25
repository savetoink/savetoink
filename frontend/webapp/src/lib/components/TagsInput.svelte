<script lang="ts">
	interface Props {
		value: string[];
		onChange: (tags: string[]) => void;
		disabled?: boolean;
		maxTags?: number;
		maxTagLength?: number;
		error?: string | null;
	}

	let {
		value = [],
		onChange,
		disabled = false,
		maxTags = 10,
		maxTagLength = 50,
		error = null
	}: Props = $props();

	let input = $state('');

	function addTag(tag: string) {
		const normalized = tag.trim().toLowerCase();
		if (!normalized) return;

		if (value.includes(normalized)) {
			return; // Duplicate
		}

		if (value.length >= maxTags) {
			return; // Max reached
		}

		if (normalized.length > maxTagLength) {
			return; // Too long
		}

		onChange([...value, normalized]);
		input = '';
	}

	function removeTag(tag: string) {
		onChange(value.filter((t) => t !== tag));
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' || event.key === ',') {
			event.preventDefault();
			addTag(input);
		}
		if (event.key === 'Backspace' && input === '' && value.length > 0) {
			removeTag(value[value.length - 1]);
		}
	}

	let tagCountLabel = $derived(`${value.length}/${maxTags}`);
</script>

<div class="tags-input">
	<ul class="tags-list">
		{#each value as tag (tag)}
			<li class="tag">
				<ins>#{tag}</ins>
				<button
					type="button"
					class="remove-tag"
					onclick={() => removeTag(tag)}
					{disabled}
					aria-label={`Remove tag ${tag}`}
				>
					×
				</button>
			</li>
		{/each}
	</ul>

	<input
		type="text"
		bind:value={input}
		placeholder={value.length === 0 ? 'Add tags...' : ''}
		disabled={disabled || value.length >= maxTags}
		onkeydown={handleKeydown}
	/>

	<small class="tag-count">
		{tagCountLabel}
		{#if error}
			<span class="error">{error}</span>
		{/if}
	</small>
</div>

<style>
	.tags-input {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.tags-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		padding: 0;
		margin: 0;
		list-style: none;
	}

	.tag {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		background-color: var(--pico-primary-background);
		color: var(--pico-color);
		border-radius: var(--pico-border-radius);
		padding: 0.25rem 0.5rem;
	}

	.tag ins {
		text-decoration: none;
	}

	.remove-tag {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1rem;
		height: 1rem;
		padding: 0;
		margin-left: 0.25rem;
		border: none;
		background: transparent;
		color: inherit;
		cursor: pointer;
		font-size: 1.2rem;
		line-height: 1;
		border-radius: 50%;
	}

	.remove-tag:hover:not(:disabled) {
		background-color: rgba(0, 0, 0, 0.1);
	}

	.remove-tag:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	input {
		width: 100%;
	}

	.tag-count {
		color: var(--muted-color);
		font-size: 0.875rem;
	}

	.tag-count .error {
		color: var(--pico-del-color);
		margin-left: 0.5rem;
	}
</style>
