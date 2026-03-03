export const API_URL = import.meta.env.VITE_SAVETOINK_API_URL;

const send = async (path: string, token: string): Promise<Response> => {
  const url = `${API_URL}${path}`;
  const options: RequestInit = {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  };
  return await fetch(url, options);
};

export interface ProfileResponse {
  ok: boolean;
  status: number;
  profile?: {
    account: string;
    email: string;
    device_email: string;
    auto_send: boolean;
  };
}

export const getProfile = async (token: string): Promise<ProfileResponse> => {
  const response = await send("/v1/user/profile", token);
  const result: ProfileResponse = {
    ok: response.ok,
    status: response.status,
  };

  if (response.ok) {
    try {
      const data = await response.json();
      result.profile = data;
    } catch (error) {
      console.error("failed to parse profile response:", error);
    }
  }

  return result;
};

export interface CreateArticleResponse {
  id: string;
  title: string;
  url: string;
}

export const createArticle = async (url: string, token: string): Promise<CreateArticleResponse> => {
  const response = await fetch(`${API_URL}/v1/articles/`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ url }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || "failed to create article");
  }

  return response.json();
};
