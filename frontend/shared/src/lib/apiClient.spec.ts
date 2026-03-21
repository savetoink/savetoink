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

	it('should return SendsResponse with quota information for Auth0 users', async () => {
		const { createApiClient } = await import('./apiClient.js');

		mockFetch.mockResolvedValue({
			ok: true,
			text: async () =>
				JSON.stringify({
					total_sends: 100,
					current_sends: 15,
					max_sends_per_period: 50,
					period_days: 30,
					remaining_sends: 35,
					period_reset_date: '2024-04-01T00:00:00Z'
				})
		});

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch
		});

		const result = await client.getSends('test-token');

		expect(result).toEqual({
			total_sends: 100,
			current_sends: 15,
			max_sends_per_period: 50,
			period_days: 30,
			remaining_sends: 35,
			period_reset_date: '2024-04-01T00:00:00Z'
		});
		expect(mockFetch).toHaveBeenCalledWith(
			'https://api.example.com/v1/sends',
			expect.objectContaining({
				method: 'GET',
				headers: expect.objectContaining({
					'Authorization': 'Bearer test-token'
				})
			})
		);
	});

	it('should return SendsResponseNoLimits for shared API key users', async () => {
		const { createApiClient } = await import('./apiClient.js');

		mockFetch.mockResolvedValue({
			ok: true,
			text: async () => JSON.stringify({ total_sends: 42 })
		});

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch
		});

		const result = await client.getSends('test-token');

		expect(result).toEqual({ total_sends: 42 });
	});

	it('should throw ApiError when getSends fails with 401', async () => {
		const { createApiClient, ApiError } = await import('./apiClient.js');

		mockFetch.mockResolvedValue({
			ok: false,
			status: 401,
			statusText: 'Unauthorized',
			json: async () => ({ error: 'Invalid token' })
		});

		const client = createApiClient({
			baseUrl: 'https://api.example.com',
			fetch: mockFetch
		});

		await expect(client.getSends('test-token')).rejects.toThrow(ApiError);
		await expect(client.getSends('test-token')).rejects.toMatchObject({
			status: 401,
			message: 'Invalid token'
		});
	});
});
