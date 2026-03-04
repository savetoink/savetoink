export interface Article {
  account: string;
  id: string;
  url: string;
  createdAt: string;
  title?: string;
  content?: string;
  author?: string;
  siteName?: string;
  sourceDomain?: string;
  excerpt?: string;
  imageUrl?: string;
  contentType?: string;
  language?: string;
  error?: string;
  wordCount?: number;
  readingTimeMinutes?: number;
  publishedAt?: string;
  favorite?: boolean;
}

export interface BounceInfo {
  timestamp: string;
  error: string;
}

export interface UserProfile {
  account: string;
  email: string;
  deviceEmail?: string;
  autoSend?: boolean;
  bouncedEmails?: Record<string, BounceInfo>;
}

export interface Send {
  account: string;
  articleId: string;
  sentAt: string;
  title: string;
  destEmail: string;
  status: string;
  senderEmail: string;
  messageId?: string;
  provider: string;
  errorResponse?: string;
}

export interface ApiErrorResponse {
  error: string;
}
