// @vitest-environment happy-dom
import { describe, expect, it } from 'vitest';
import { isEditableElement } from './isEditableElement';

describe('isEditableElement', () => {
	it('should identify text input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'text';
		expect(isEditableElement(input)).toBe(true);
	});

	it('should identify url input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'url';
		expect(isEditableElement(input)).toBe(true);
	});

	it('should identify email input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'email';
		expect(isEditableElement(input)).toBe(true);
	});

	it('should identify password input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'password';
		expect(isEditableElement(input)).toBe(true);
	});

	it('should identify search input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'search';
		expect(isEditableElement(input)).toBe(true);
	});

	it('should identify tel input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'tel';
		expect(isEditableElement(input)).toBe(true);
	});

	it('should identify number input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'number';
		expect(isEditableElement(input)).toBe(true);
	});

	it('should not identify checkbox input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'checkbox';
		expect(isEditableElement(input)).toBe(false);
	});

	it('should not identify radio input elements as editable', () => {
		const input = document.createElement('input');
		input.type = 'radio';
		expect(isEditableElement(input)).toBe(false);
	});

	it('should identify default input elements as editable (text type by default)', () => {
		const input = document.createElement('input');
		expect(isEditableElement(input)).toBe(true);
	});

	it('should identify textarea elements as editable', () => {
		const textarea = document.createElement('textarea');
		expect(isEditableElement(textarea)).toBe(true);
	});

	it('should identify select elements as editable', () => {
		const select = document.createElement('select');
		expect(isEditableElement(select)).toBe(true);
	});

	it('should identify contenteditable elements as editable', () => {
		const div = document.createElement('div');
		div.contentEditable = 'true';
		expect(isEditableElement(div)).toBe(true);
	});

	it('should not identify non-editable div elements as editable', () => {
		const div = document.createElement('div');
		expect(isEditableElement(div)).toBe(false);
	});

	it('should not identify button elements as editable', () => {
		const button = document.createElement('button');
		expect(isEditableElement(button)).toBe(false);
	});

	it('should not identify anchor elements as editable', () => {
		const anchor = document.createElement('a');
		expect(isEditableElement(anchor)).toBe(false);
	});

	it('should handle null target safely', () => {
		expect(isEditableElement(null)).toBe(false);
	});

	it('should handle non-HTMLElement target safely', () => {
		expect(isEditableElement('string' as unknown as EventTarget)).toBe(false);
	});

	it('should not identify contenteditable="false" elements as editable', () => {
		const div = document.createElement('div');
		div.contentEditable = 'false';
		expect(isEditableElement(div)).toBe(false);
	});

	it('should handle case-insensitive input types', () => {
		const input = document.createElement('input');
		input.type = 'TEXT';
		expect(isEditableElement(input)).toBe(true);
	});
});
