import { describe, it, expect, vi, beforeEach } from 'vitest';
import { toggleFavorite, deleteArticle, sendArticle } from '@savetoink/shared';

describe('form actions', () => {
	let mockForm: HTMLFormElement;

	beforeEach(() => {
		mockForm = {
			requestSubmit: vi.fn()
		} as unknown as HTMLFormElement;
		global.window = {
			confirm: vi.fn()
		} as unknown as Window & typeof globalThis;
	});

	describe('toggleFavorite', () => {
		it('should call requestSubmit on the form', () => {
			toggleFavorite(mockForm);
			expect(mockForm.requestSubmit).toHaveBeenCalledTimes(1);
		});
	});

	describe('deleteArticle', () => {
		it('should call requestSubmit when user confirms', () => {
			vi.spyOn(window, 'confirm').mockReturnValue(true);
			deleteArticle(mockForm);
			expect(window.confirm).toHaveBeenCalledWith('Are you sure you want to delete this article?');
			expect(mockForm.requestSubmit).toHaveBeenCalledTimes(1);
		});

		it('should not call requestSubmit when user cancels', () => {
			vi.spyOn(window, 'confirm').mockReturnValue(false);
			deleteArticle(mockForm);
			expect(window.confirm).toHaveBeenCalledWith('Are you sure you want to delete this article?');
			expect(mockForm.requestSubmit).not.toHaveBeenCalled();
		});
	});

	describe('sendArticle', () => {
		it('should call requestSubmit on the form', () => {
			sendArticle(mockForm);
			expect(mockForm.requestSubmit).toHaveBeenCalledTimes(1);
		});
	});
});
