import type {
  Article,
  CreateArticleResponse,
  ExchangeCodeResponse,
  Send,
  UserProfile,
} from "../types";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export interface GetArticlesParams {
  page?: number;
  pageSize?: number;
  favorite?: boolean;
}

export interface ArticlesResponse {
  articles: Article[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ApiClient {
  getProfile(token: string): Promise<UserProfile>;
  getArticles(params: GetArticlesParams, token: string): Promise<ArticlesResponse>;
  getArticle(id: string, token: string): Promise<Article>;
  createArticle(url: string, tags: string[] | null, token: string): Promise<CreateArticleResponse>;
  sendArticle(id: string, token: string): Promise<void>;
  favoriteArticle(id: string, token: string): Promise<void>;
  deleteArticle(id: string, token: string): Promise<void>;
  updateDevice(deviceEmail: string, autoSend: boolean, token: string): Promise<void>;
  deleteDevice(token: string): Promise<void>;
  exchangeCodeForToken(code: string, redirectUri: string): Promise<ExchangeCodeResponse>;
  getSends(articleId: string, token: string): Promise<Send[]>;
}

export interface ApiClientOptions {
  baseUrl: string;
  fetch?: typeof globalThis.fetch;
}

export function createApiClient({ baseUrl, fetch }: ApiClientOptions): ApiClient {
  const fetchFn = fetch || globalThis.fetch;

  async function request<T>(
    method: string,
    path: string,
    token?: string,
    body?: unknown
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetchFn(`${baseUrl}${path}`, {
      method,
      headers,
      ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
    });

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new ApiError(res.status, err.error ?? err.message ?? res.statusText);
    }

    const text = await res.text();
    return text ? JSON.parse(text) : ({} as T);
  }

  return {
    getProfile: (token: string) =>
      request<UserProfile>("GET", "/v1/user/profile", token),

    getArticles: (params: GetArticlesParams, token: string) => {
      const queryParams = new URLSearchParams();
      if (params.page !== undefined) {
        queryParams.set("page", params.page.toString());
      }
      if (params.pageSize !== undefined) {
        queryParams.set("page_size", params.pageSize.toString());
      }
      if (params.favorite !== undefined) {
        queryParams.set("favorite", params.favorite.toString());
      }
      const path = `/v1/articles${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;
      return request<ArticlesResponse>("GET", path, token);
    },

    getArticle: (id: string, token: string) =>
      request<Article>("GET", `/v1/articles/${id}`, token),

    createArticle: (url: string, tags: string[] | null, token: string) =>
      request<CreateArticleResponse>("POST", "/v1/articles", token, {
        url,
        ...(tags && { tags }),
      }),

    sendArticle: (id: string, token: string) =>
      request<void>("POST", `/v1/articles/${id}/send`, token),

    favoriteArticle: (id: string, token: string) =>
      request<void>("PUT", `/v1/articles/${id}/favorite`, token),

    deleteArticle: (id: string, token: string) =>
      request<void>("DELETE", `/v1/articles/${id}`, token),

    updateDevice: (deviceEmail: string, autoSend: boolean, token: string) =>
      request<void>("PUT", "/v1/devices", token, {
        device_email: deviceEmail,
        auto_send: autoSend,
      }),

    deleteDevice: (token: string) =>
      request<void>("DELETE", "/v1/devices", token),

    exchangeCodeForToken: (code: string, redirectUri: string) =>
      request<ExchangeCodeResponse>("POST", "/v1/auth/token", undefined, {
        code,
        redirect_uri: redirectUri,
      }),

    getSends: (articleId: string, token: string) =>
      request<Send[]>("GET", `/v1/articles/${articleId}/sends`, token),
  };
}
