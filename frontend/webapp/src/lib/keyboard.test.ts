import { describe, expect, it } from 'vitest';
import { BASE_BINDINGS, LIST_BINDINGS, DETAIL_BINDINGS } from '@savetoink/shared';

describe('keyboard bindings', () => {
	it('should have all expected base bindings', () => {
		expect(Object.keys(BASE_BINDINGS)).toEqual(['f', 'd', 's', 'n', 'h', 'a']);
	});

	it('should have correct descriptions for base bindings', () => {
		expect(BASE_BINDINGS.f.description).toBe('Toggle favorite');
		expect(BASE_BINDINGS.d.description).toBe('Delete article');
		expect(BASE_BINDINGS.s.description).toBe('Send to device');
		expect(BASE_BINDINGS.n.description).toBe('New article');
		expect(BASE_BINDINGS.h.description).toBe('Go home');
		expect(BASE_BINDINGS.a.description).toBe('Account page');
	});

	it('should have correct categories for base bindings', () => {
		expect(BASE_BINDINGS.f.category).toBe('action');
		expect(BASE_BINDINGS.d.category).toBe('action');
		expect(BASE_BINDINGS.s.category).toBe('action');
		expect(BASE_BINDINGS.n.category).toBe('navigation');
		expect(BASE_BINDINGS.h.category).toBe('navigation');
		expect(BASE_BINDINGS.a.category).toBe('navigation');
	});

	it('should include all base bindings in list bindings', () => {
		Object.keys(BASE_BINDINGS).forEach((key) => {
			expect(LIST_BINDINGS[key]).toBeDefined();
		});
	});

	it('should have list-specific bindings', () => {
		expect(LIST_BINDINGS.k).toBeDefined();
		expect(LIST_BINDINGS.j).toBeDefined();
		expect(LIST_BINDINGS.Enter).toBeDefined();
	});

	it('should have correct descriptions for list bindings', () => {
		expect(LIST_BINDINGS.k.description).toBe('Previous article');
		expect(LIST_BINDINGS.j.description).toBe('Next article');
		expect(LIST_BINDINGS.Enter.description).toBe('Open article');
	});

	it('should have correct category for list bindings', () => {
		expect(LIST_BINDINGS.k.category).toBe('list');
		expect(LIST_BINDINGS.j.category).toBe('list');
		expect(LIST_BINDINGS.Enter.category).toBe('list');
	});

	it('should include all base bindings in detail bindings', () => {
		Object.keys(BASE_BINDINGS).forEach((key) => {
			expect(DETAIL_BINDINGS[key]).toBeDefined();
		});
	});

	it('should have detail-specific bindings', () => {
		expect(DETAIL_BINDINGS.Escape).toBeDefined();
	});

	it('should have correct descriptions for detail bindings', () => {
		expect(DETAIL_BINDINGS.Escape.description).toBe('Back to list');
	});

	it('should have correct category for detail bindings', () => {
		expect(DETAIL_BINDINGS.Escape.category).toBe('navigation');
	});
});
