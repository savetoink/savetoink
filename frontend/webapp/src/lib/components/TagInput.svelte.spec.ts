import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import TagInput from './TagInput.svelte';

describe('TagInput', () => {
	it('renders empty tag list', () => {
		render(TagInput, { tags: [] });
		expect(screen.getByText('No tags yet')).toBeTruthy();
	});

	it('renders existing tags', () => {
		render(TagInput, { tags: ['reading', 'tech'] });
		expect(screen.getByText('reading')).toBeTruthy();
		expect(screen.getByText('tech')).toBeTruthy();
	});

	it('does not show add tags input when onAdd is not provided', () => {
		render(TagInput, { tags: ['reading'] });
		expect(screen.queryByPlaceholderText('reading, tech, tutorial')).toBeNull();
	});

	it('shows add tags input when onAdd is provided', () => {
		render(TagInput, { tags: ['reading'], onAdd: () => Promise.resolve() });
		expect(screen.getByPlaceholderText('reading, tech, tutorial')).toBeTruthy();
	});

	it('does not show remove buttons when onRemove is not provided', () => {
		const { container } = render(TagInput, { tags: ['reading'], onAdd: () => Promise.resolve() });
		const removeButtons = container.querySelectorAll('.tag-remove');
		expect(removeButtons.length).toBe(0);
	});

	it('shows remove buttons when onRemove is provided', () => {
		const { container } = render(TagInput, {
			tags: ['reading'],
			onAdd: () => Promise.resolve(),
			onRemove: () => Promise.resolve()
		});
		const removeButtons = container.querySelectorAll('.tag-remove');
		expect(removeButtons.length).toBe(1);
	});
});
