import { createApiClient } from "@savetoink/shared";

const apiClient = createApiClient({ baseUrl: import.meta.env.PUBLIC_API_URL });

export const API_URL = import.meta.env.PUBLIC_API_URL;

export const getProfile = apiClient.getProfile;
export const createArticle = apiClient.createArticle;
export const sendArticle = apiClient.sendArticle;
export const exchangeCodeForToken = apiClient.exchangeCodeForToken;
