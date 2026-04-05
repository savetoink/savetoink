<script lang="ts">
	import { enhance } from '$app/forms';
	import { tick } from 'svelte';
	import Tags from './Tags.svelte';
	import { MAX_TAGS, MAX_TAG_LENGTH } from '@savetoink/shared';

	let {
		tags,
		articleId,
		editable = false,
		showAddTag = false,
		onClose,
		onInputReady
	}: {
		tags: string[] | undefined;
		articleId?: string;
		editable?: boolean;
		showAddTag?: boolean;
		onClose?: () => void;
		onInputReady?: (element: HTMLInputElement) => void;
	} = $props();

	let tagsInput = $state('');
	let tagsError = $state<string | null>(null);
	let inputElement = $state<HTMLInputElement>();

	// Focus input when showAddTag becomes true
	$effect(() => {
		if (showAddTag) {
			tick().then(() => {
				inputElement?.focus();
			});
		}
	});

	// Call onInputReady when input element is available
	$effect(() => {
		if (inputElement && onInputReady) {
			onInputReady(inputElement);
		}
	});

	const canAddTags = $derived(editable && articleId);
	const currentTagCount = $derived(tags?.length || 0);
	const parsedTags = $derived(() => {
		if (!tagsInput) return [];
		return tagsInput
			.split(',')
			.map((tag) => tag.trim())
			.filter((tag) => tag.length > 0);
	});

	function validateTags(tags: string[]): string | null {
		if (tags.length > MAX_TAGS) {
			return `Maximum ${MAX_TAGS} tags allowed per article`;
		}
		for (const tag of tags) {
			if (tag.length > MAX_TAG_LENGTH) {
				return `Tag "${tag}" exceeds maximum length of ${MAX_TAG_LENGTH} characters`;
			}
		}
		return null;
	}

	async function handleEnhance() {
		tagsError = null;

		const tags = parsedTags();
		const validationError = validateTags(tags);
		if (validationError) {
			tagsError = validationError;
			return;
		}

		return async ({
			update
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
		}) => {
			await update();
			tagsInput = '';
			onClose?.();
			tagsError = null;
		};
	}

	function handleCancel() {
		onClose?.();
		tagsInput = '';
		tagsError = null;
	}
</script>

<Tags {tags} {articleId} {editable} />

{#if canAddTags && showAddTag}
	<form method="POST" action="?/addTags" use:enhance={handleEnhance}>
		<label>
			Tags
			<input
				bind:this={inputElement}
				type="text"
				name="tags"
				bind:value={tagsInput}
				placeholder="reading, tech, tutorial"
				aria-describedby="tags-help"
			/>
			<small id="tags-help">
				Comma-separated tags (max {MAX_TAGS} total, {MAX_TAG_LENGTH} characters each). Current:
				{currentTagCount}/{MAX_TAGS}. {parsedTags().length > 0 &&
					`Adding: ${parsedTags().length} more.`}
			</small>
		</label>
		<div>
			<button type="submit">Add</button>
			<button type="button" onclick={handleCancel}>Cancel</button>
		</div>
		{#if tagsError}
			<p role="alert" style="color: red">{tagsError}</p>
		{/if}
	</form>
{/if}
