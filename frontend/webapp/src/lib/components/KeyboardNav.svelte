<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { KeyBindingMap } from '@savetoink/shared';
	import { HELP_KEY } from '@savetoink/shared';

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
			<h2>Keyboard Shortcuts</h2>
			<button class="close-btn" onclick={() => closeHelpModal()} aria-label="Close">
				&times;
			</button>
		</header>
		<div class="bindings-container">
			{#each Object.entries(groupBindingsByCategory(bindings, allEnabledKeys)) as [category, categoryBindings] (category)}
				<div class="category">
					<h3>{category}</h3>
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
		<footer>
			<button onclick={() => closeHelpModal()}>Close</button>
			<small>Press {HELP_KEY} or Escape to close</small>
		</footer>
	</div>
</dialog>

<style>
	.help-modal {
		max-width: 500px;
		width: 90vw;
		border: none;
		border-radius: 8px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
		padding: 0;
	}

	.help-modal::backdrop {
		background: rgba(0, 0, 0, 0.5);
	}

	.modal-content {
		padding: 1.5rem;
		background: white;
		color: inherit;
	}

	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	header h2 {
		margin: 0;
		font-size: 1.5rem;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 2rem;
		line-height: 1;
		cursor: pointer;
		padding: 0 0.5rem;
	}

	.bindings-container {
		max-height: 60vh;
		overflow-y: auto;
		margin-bottom: 1rem;
	}

	.category {
		margin-bottom: 1rem;
	}

	.category h3 {
		margin: 0 0 0.5rem 0;
		font-size: 1rem;
		text-transform: capitalize;
		color: #666;
	}

	.category ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.category li {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.5rem 0;
	}

	kbd {
		background: #f4f4f4;
		border: 1px solid #ccc;
		border-radius: 4px;
		padding: 0.25rem 0.5rem;
		font-family: monospace;
		font-size: 0.9rem;
		min-width: 2.5rem;
		text-align: center;
	}

	footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		border-top: 1px solid #eee;
		padding-top: 1rem;
	}

	footer small {
		color: #666;
	}
</style>
