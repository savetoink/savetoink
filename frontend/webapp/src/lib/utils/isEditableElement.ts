export function isEditableElement(element: EventTarget | null): boolean {
	if (!(element instanceof HTMLElement)) {
		return false;
	}
	const tagName = element.tagName.toLowerCase();

	// Check for contenteditable elements
	if (element.isContentEditable) {
		return true;
	}

	// Check for text-based input types (excluding checkboxes, radio buttons, etc.)
	if (tagName === 'input') {
		const inputElement = element as HTMLInputElement;
		const inputType = inputElement.type.toLowerCase();
		const textInputTypes = ['text', 'url', 'email', 'password', 'search', 'tel', 'number'];
		return textInputTypes.includes(inputType);
	}

	// Check for textarea and select elements
	return tagName === 'textarea' || tagName === 'select';
}
