import type {
  UserProfile,
  Article,
  ArticleResponse,
  AuthTokenExchangeResponse,
  AuthTokenExchangeRequest,
  ArticleRequest,
  DeviceRequest,
  SendsResponse,
  SendsResponseNoLimits,
} from "../types";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export interface ApiClient {
  getProfile(token: string): Promise<UserProfile>;
  getArticles(
    params: {
      page?: number;
      page_size?: number;
      favorite?: boolean;
    },
    token: string,
  ): Promise<{
    articles: Article[];
    page: number;
    page_size: number;
    total: number;
    has_more: boolean;
  }>;
  getArticle(id: string, token: string): Promise<Article>;
  createArticle(
    url: string,
    sendToDevice: boolean,
    token: string,
  ): Promise<ArticleResponse>;
  sendArticle(id: string, token: string): Promise<void>;
  favoriteArticle(id: string, token: string): Promise<void>;
  deleteArticle(id: string, token: string): Promise<void>;
  updateDevice(
    deviceEmail: string,
    autoSend: boolean,
    token: string,
  ): Promise<void>;
  deleteDevice(token: string): Promise<void>;
  exchangeCodeForToken(
    code: string,
    redirectUri: string,
  ): Promise<AuthTokenExchangeResponse>;
  getSends(token: string): Promise<SendsResponse | SendsResponseNoLimits>;
}

export interface ApiClientOptions {
  baseUrl: string;
  fetch?: typeof globalThis.fetch;
  userAgent?: string;
  xForwardedFor?: string;
  cloudFlareRay?: string;
}

export function createApiClient({
  baseUrl,
  fetch,
  userAgent,
  xForwardedFor,
  cloudFlareRay,
}: ApiClientOptions): ApiClient {
  const fetchFn = fetch || globalThis.fetch;

  async function request<T>(
    method: string,
    path: string,
    token?: string,
    body?: unknown,
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    if (userAgent) {
      headers["User-Agent"] = userAgent;
    }

    if (xForwardedFor) {
      headers["X-Forwarded-For"] = xForwardedFor;
    }

    if (cloudFlareRay) {
      headers["CF-Ray"] = cloudFlareRay;
    }

    let res: Response;
    try {
      res = await fetchFn(`${baseUrl}${path}`, {
        method,
        headers,
        ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
      });
    } catch (e) {
      throw new ApiError(
        503,
        e instanceof Error ? e.message : "Network error: Failed to connect to the server",
      );
    }

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new ApiError(
        res.status,
        err.error ?? err.message ?? res.statusText,
      );
    }

    const text = await res.text();
    return text ? JSON.parse(text) : ({} as T);
  }

  return {
    getProfile: (token: string) =>
      request<UserProfile>("GET", "/v1/user/profile", token),

    getArticles: (
      params: {
        page?: number;
        page_size?: number;
        favorite?: boolean;
      },
      token: string,
    ) => {
      const queryParams = new URLSearchParams();
      if (params.page !== undefined) {
        queryParams.set("page", params.page.toString());
      }
      if (params.page_size !== undefined) {
        queryParams.set("page_size", params.page_size.toString());
      }
      if (params.favorite !== undefined) {
        queryParams.set("favorite", params.favorite.toString());
      }
      const path = `/v1/articles${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;
      return request<{
        articles: Article[];
        page: number;
        page_size: number;
        total: number;
        has_more: boolean;
      }>("GET", path, token);
    },

    getArticle: (id: string, token: string) =>
      request<Article>("GET", `/v1/articles/${id}`, token),

    createArticle: (url: string, sendToDevice: boolean, token: string) =>
      request<ArticleResponse>("POST", "/v1/articles", token, {
        url,
        send_on_complete: sendToDevice,
      } as ArticleRequest),

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
      } as DeviceRequest),

    deleteDevice: (token: string) =>
      request<void>("DELETE", "/v1/devices", token),

    exchangeCodeForToken: (code: string, redirectUri: string) =>
      request<AuthTokenExchangeResponse>("POST", "/v1/auth/token", undefined, {
        code,
        redirect_uri: redirectUri,
      } as AuthTokenExchangeRequest),

    getSends: (token: string) =>
      request<SendsResponse | SendsResponseNoLimits>("GET", "/v1/sends", token),
  };
}
