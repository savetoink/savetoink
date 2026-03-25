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

	describe('tag methods', () => {
		it('should add tags to article', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () => JSON.stringify({ tags: ['tech', 'reading', 'tutorial'] })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			const result = await client.addTags('article-123', ['reading', 'tutorial'], 'test-token');

			expect(result).toEqual({ tags: ['tech', 'reading', 'tutorial'] });
			expect(mockFetch).toHaveBeenCalledWith(
				'https://api.example.com/v1/articles/article-123/tags',
				expect.objectContaining({
					method: 'POST',
					headers: expect.objectContaining({
						'Authorization': 'Bearer test-token',
						'Content-Type': 'application/json'
					}),
					body: JSON.stringify({ tags: ['reading', 'tutorial'] })
				})
			);
		});

		it('should set all tags for article', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () => JSON.stringify({ tags: ['tech', 'tutorial'] })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			const result = await client.setTags('article-123', ['tech', 'tutorial'], 'test-token');

			expect(result).toEqual({ tags: ['tech', 'tutorial'] });
			expect(mockFetch).toHaveBeenCalledWith(
				'https://api.example.com/v1/articles/article-123/tags',
				expect.objectContaining({
					method: 'PUT',
					headers: expect.objectContaining({
						'Authorization': 'Bearer test-token',
						'Content-Type': 'application/json'
					}),
					body: JSON.stringify({ tags: ['tech', 'tutorial'] })
				})
			);
		});

		it('should set empty array to remove all tags', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () => JSON.stringify({ tags: [] })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			const result = await client.setTags('article-123', [], 'test-token');

			expect(result).toEqual({ tags: [] });
			expect(mockFetch).toHaveBeenCalledWith(
				'https://api.example.com/v1/articles/article-123/tags',
				expect.objectContaining({
					method: 'PUT',
					body: JSON.stringify({ tags: [] })
				})
			);
		});

		it('should get tags for article', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () => JSON.stringify({ tags: ['tech', 'reading'] })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			const result = await client.getTags('article-123', 'test-token');

			expect(result).toEqual({ tags: ['tech', 'reading'] });
			expect(mockFetch).toHaveBeenCalledWith(
				'https://api.example.com/v1/articles/article-123/tags',
				expect.objectContaining({
					method: 'GET',
					headers: expect.objectContaining({
						'Authorization': 'Bearer test-token'
					})
				})
			);
		});

		it('should remove tags from article', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () => JSON.stringify({ tags: ['tech'] })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			const result = await client.removeTags('article-123', ['reading', 'tutorial'], 'test-token');

			expect(result).toEqual({ tags: ['tech'] });
			expect(mockFetch).toHaveBeenCalledWith(
				'https://api.example.com/v1/articles/article-123/tags',
				expect.objectContaining({
					method: 'DELETE',
					headers: expect.objectContaining({
						'Authorization': 'Bearer test-token',
						'Content-Type': 'application/json'
					}),
					body: JSON.stringify({ tags: ['reading', 'tutorial'] })
				})
			);
		});

		it('should get all tags for account', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () =>
					JSON.stringify({ tags: ['tech', 'reading', 'tutorial', 'news', 'dev'] })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			const result = await client.getAllTags('test-token');

			expect(result).toEqual({ tags: ['tech', 'reading', 'tutorial', 'news', 'dev'] });
			expect(mockFetch).toHaveBeenCalledWith(
				'https://api.example.com/v1/tags',
				expect.objectContaining({
					method: 'GET',
					headers: expect.objectContaining({
						'Authorization': 'Bearer test-token'
					})
				})
			);
		});

		it('should include tag parameter in getArticles URL', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () => JSON.stringify({
					articles: [],
					page: 1,
					page_size: 10,
					total: 0,
					has_more: false
				})
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			await client.getArticles({ tag: 'tech', page: 1 }, 'test-token');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('tag=tech'),
				expect.objectContaining({
					method: 'GET',
					headers: expect.objectContaining({
						'Authorization': 'Bearer test-token'
					})
				})
			);
		});

		it('should handle tag filtering with other parameters', async () => {
			const { createApiClient } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: true,
				text: async () => JSON.stringify({
					articles: [],
					page: 2,
					page_size: 20,
					total: 0,
					has_more: false
				})
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			await client.getArticles({ tag: 'reading', favorite: true, page: 2, page_size: 20 }, 'test-token');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('tag=reading'),
				expect.objectContaining({
					method: 'GET'
				})
			);
		});

		it('should throw ApiError when addTags fails with 400', async () => {
			const { createApiClient, ApiError } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: false,
				status: 400,
				statusText: 'Bad Request',
				json: async () => ({ error: 'Maximum 10 tags allowed per article' })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			await expect(client.addTags('article-123', Array(11).fill('tag'), 'test-token')).rejects.toThrow(
				ApiError
			);
			await expect(client.addTags('article-123', Array(11).fill('tag'), 'test-token')).rejects.toMatchObject(
				{
					status: 400,
					message: 'Maximum 10 tags allowed per article'
				}
			);
		});

		it('should throw ApiError when setTags fails with 404', async () => {
			const { createApiClient, ApiError } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: false,
				status: 404,
				statusText: 'Not Found',
				json: async () => ({ error: 'Article not found' })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			await expect(client.setTags('nonexistent-article', ['tag'], 'test-token')).rejects.toThrow(ApiError);
			await expect(client.setTags('nonexistent-article', ['tag'], 'test-token')).rejects.toMatchObject({
				status: 404,
				message: 'Article not found'
			});
		});

		it('should throw ApiError when getTags fails with 401', async () => {
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

			await expect(client.getTags('article-123', 'invalid-token')).rejects.toThrow(ApiError);
			await expect(client.getTags('article-123', 'invalid-token')).rejects.toMatchObject({
				status: 401,
				message: 'Invalid token'
			});
		});

		it('should throw ApiError when removeTags fails with 500', async () => {
			const { createApiClient, ApiError } = await import('./apiClient.js');

			mockFetch.mockResolvedValue({
				ok: false,
				status: 500,
				statusText: 'Internal Server Error',
				json: async () => ({ error: 'Database error' })
			});

			const client = createApiClient({
				baseUrl: 'https://api.example.com',
				fetch: mockFetch
			});

			await expect(client.removeTags('article-123', ['tag'], 'test-token')).rejects.toThrow(ApiError);
			await expect(client.removeTags('article-123', ['tag'], 'test-token')).rejects.toMatchObject({
				status: 500,
				message: 'Database error'
			});
		});

		it('should throw ApiError when getAllTags fails with 500', async () => {
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

			await expect(client.getAllTags('test-token')).rejects.toThrow(ApiError);
			await expect(client.getAllTags('test-token')).rejects.toMatchObject({
				status: 500,
				message: 'Server error'
			});
		});
	});
});
