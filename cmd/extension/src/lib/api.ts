import type { SendArticleResponse } from './types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://api.saveto.ink/v1';

export async function sendArticle(url: string, accessToken: string): Promise<SendArticleResponse> {
  const response = await fetch(`${API_BASE_URL}/articles`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ url }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to send article: ${response.status} ${errorText}`);
  }

  return response.json();
}
