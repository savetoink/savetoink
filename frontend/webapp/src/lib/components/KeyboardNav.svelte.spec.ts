import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import KeyboardNav from './KeyboardNav.svelte';
import { LIST_BINDINGS } from '@savetoink/shared';

describe('KeyboardNav.svelte', () => {
	it('should render without errors', () => {
		const mockCallbacks = { f: vi.fn(), n: vi.fn() };
		expect(() =>
			render(KeyboardNav, { bindings: LIST_BINDINGS, callbacks: mockCallbacks })
		).not.toThrow();
	});

	it('should render with enabled keys filter', () => {
		const mockBindings = {
			a: { description: 'Test A', category: 'action' },
			b: { description: 'Test B', category: 'action' }
		};
		const mockCallbacks = { a: vi.fn(), b: vi.fn() };
		expect(() =>
			render(KeyboardNav, {
				bindings: mockBindings,
				callbacks: mockCallbacks,
				enabledKeys: ['a']
			})
		).not.toThrow();
	});
});
