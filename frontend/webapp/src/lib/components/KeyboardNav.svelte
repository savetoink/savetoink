<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { KeyBindingMap } from '@savetoink/shared';
	import { HELP_KEY } from '@savetoink/shared';
	import { isEditableElement } from '$lib/utils/isEditableElement';

	let {
		bindings,
		callbacks,
		enabledKeys
	}: {
		bindings: KeyBindingMap;
		callbacks: Record<string, (e: KeyboardEvent) => void | Promise<void>>;
		enabledKeys?: string[];
	} = $props();

	let showHelpModal = $state(false);
	let dialogElement: HTMLDialogElement;

	const allEnabledKeys = $derived(enabledKeys || Object.keys(bindings));

	function handleKeydown(e: KeyboardEvent) {
		// Disable keyboard shortcuts when user is typing in an input field
		if (isEditableElement(e.target)) {
			return;
		}

		if (showHelpModal && (e.key === 'Escape' || e.key === HELP_KEY)) {
			e.preventDefault();
			closeHelpModal();
			return;
		}

		if (showHelpModal) {
			return;
		}

		if (e.key === HELP_KEY) {
			e.preventDefault();
			openHelpModal();
			return;
		}

		if (allEnabledKeys.includes(e.key)) {
			e.preventDefault();
			const callback = callbacks[e.key];
			if (callback) {
				callback(e);
			}
		}
	}

	function openHelpModal() {
		showHelpModal = true;
		tick().then(() => {
			dialogElement.showModal();
		});
	}

	function closeHelpModal() {
		dialogElement.close();
		showHelpModal = false;
	}

	function groupBindingsByCategory(bindingsMap: KeyBindingMap, keys: string[]) {
		const grouped: Record<string, Array<{ key: string; description: string }>> = {};
		keys.forEach((key) => {
			if (bindingsMap[key]) {
				const { description, category } = bindingsMap[key];
				if (!grouped[category]) {
					grouped[category] = [];
				}
				grouped[category].push({ key, description });
			}
		});
		return grouped;
	}

	onMount(() => {
		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});
</script>

<dialog class="help-modal" bind:this={dialogElement} onclose={closeHelpModal}>
	<div class="modal-content">
		<header>
			<h1 class="center">Keyboard Shortcuts</h1>
		</header>
		<div class="categories">
			{#each Object.entries(groupBindingsByCategory(bindings, allEnabledKeys)) as [category, categoryBindings] (category)}
				<div>
					<h3 class="center"><p>{category}</p></h3>
					<ul>
						{#each categoryBindings as binding (binding.key)}
							<li>
								<kbd>{binding.key}</kbd>
								<span>{binding.description}</span>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		</div>
		<footer class="center">
			<button onclick={() => closeHelpModal()}>Close</button>
			<small>Press {HELP_KEY} or Escape to close</small>
		</footer>
	</div>
</dialog>

<style>
	.categories {
		display: flex;
		flex-wrap: wrap;
		gap: 3rem;
		margin-top: 2rem;
	}
	.center {
		text-align: center;
		text-transform: capitalize;
	}
</style>
