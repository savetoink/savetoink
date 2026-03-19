import { describe, expect, it, vi, beforeEach } from 'vitest';

describe('apiClient', () => {
	let mockFetch: ReturnType<typeof vi.fn>;

	beforeEach(() => {
		mockFetch = vi.fn();
	});

	it('should throw ApiError with status 503 when fetch fails due to network error', async () => {
		const { createApiClient, ApiError } = await import('./apiClient.js');

		mockFetch.mockRejectedValue(new Error('Failed to fetch'));

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch
		});

		await expect(client.getProfile('token')).rejects.toThrow(ApiError);
		await expect(client.getProfile('token')).rejects.toMatchObject({
			status: 503,
			message: expect.stringContaining('Network error')
		});
	});

	it('should throw ApiError with status 503 when fetch fails with unknown error', async () => {
		const { createApiClient, ApiError } = await import('./apiClient.js');

		mockFetch.mockRejectedValue('Unknown error');

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch
		});

		await expect(client.getProfile('token')).rejects.toThrow(ApiError);
		await expect(client.getProfile('token')).rejects.toMatchObject({
			status: 503,
			message: 'Network error: Failed to connect to the server'
		});
	});

	it('should throw ApiError with status from response when response is not OK', async () => {
		const { createApiClient, ApiError } = await import('./apiClient.js');

		mockFetch.mockResolvedValue({
			ok: false,
			status: 500,
			statusText: 'Internal Server Error',
			json: async () => ({ error: 'Server error' })
		});

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch
		});

		await expect(client.getProfile('token')).rejects.toThrow(ApiError);
		await expect(client.getProfile('token')).rejects.toMatchObject({
			status: 500,
			message: 'Server error'
		});
	});

	it('should throw ApiError with statusText when response json fails', async () => {
		const { createApiClient, ApiError } = await import('./apiClient.js');

		mockFetch.mockResolvedValue({
			ok: false,
			status: 404,
			statusText: 'Not Found',
			json: async () => {
				throw new Error('Invalid JSON');
			}
		});

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch
		});

		await expect(client.getProfile('token')).rejects.toThrow(ApiError);
		await expect(client.getProfile('token')).rejects.toMatchObject({
			status: 404,
			message: 'Not Found'
		});
	});

	it('should include proper headers including Authorization', async () => {
		const { createApiClient } = await import('./apiClient.js');

		mockFetch.mockResolvedValue({
			ok: true,
			text: async () => '{"account":"test"}'
		});

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch,
			userAgent: 'TestAgent',
			xForwardedFor: '1.2.3.4',
			cloudFlareRay: 'test-ray'
		});

		await client.getProfile('test-token');

		expect(mockFetch).toHaveBeenCalledWith(
			'https://api.example.com/v1/user/profile',
			expect.objectContaining({
				method: 'GET',
				headers: expect.objectContaining({
					'Content-Type': 'application/json',
					'Authorization': 'Bearer test-token',
					'User-Agent': 'TestAgent',
					'X-Forwarded-For': '1.2.3.4',
					'CF-Ray': 'test-ray'
				})
			})
		);
	});
});
