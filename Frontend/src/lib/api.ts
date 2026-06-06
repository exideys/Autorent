import type {
  ApiErrorResponse,
  ApiResponse,
  AuthResponse,
  Car,
  CarInput,
  CarRecommendationResponse,
  GoogleLoginPayload,
  ImageUploadResponse,
  LoginPayload,
  NewsArticle,
  NewsInput,
  RateUserPayload,
  RegisterPayload,
  RentalOrder,
  RentalOrderInput,
  SupportAttachment,
  SupportConversation,
  SupportMessage,
  SupportRealtimeEvent,
  TranslationResponse,
  UpdateCurrentUserPayload,
  User,
} from '../types/api';

const normalizeApiBaseUrl = (value?: string) => {
  const trimmed = (value || '').replace(/\/$/, '');

  if (!trimmed) {
    return '/api';
  }

  return trimmed.endsWith('/api') ? trimmed : `${trimmed}/api`;
};

const API_BASE_URL = normalizeApiBaseUrl(import.meta.env.VITE_API_BASE_URL);

const apiOrigin = () => API_BASE_URL.replace(/\/api$/, '');

const apiUrlForPath = (path: string) => {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }
  if (path.startsWith('/api/') && /^https?:\/\//i.test(API_BASE_URL)) {
    return `${apiOrigin()}${path}`;
  }
  if (path.startsWith('/api/')) {
    return path;
  }

  return `${API_BASE_URL}${path.startsWith('/') ? path : `/${path}`}`;
};

export const resolveApiAssetUrl = (value?: string | null, fallback = '') => {
  const trimmed = (value || '').trim();
  if (!trimmed) {
    return fallback;
  }
  if (/^(https?:|data:|blob:)/i.test(trimmed)) {
    return trimmed;
  }
  if (trimmed.startsWith('/api/') && /^https?:\/\//i.test(API_BASE_URL)) {
    return `${apiOrigin()}${trimmed}`;
  }
  return trimmed;
};

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  token?: string;
  rawResponse?: boolean;
}

const isFormDataBody = (body: unknown): body is FormData => typeof FormData !== 'undefined' && body instanceof FormData;

const buildHeaders = (options: RequestOptions) => {
  const headers = new Headers(options.headers);

  if (options.body !== undefined && !isFormDataBody(options.body) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  if (options.token) {
    headers.set('Authorization', `Bearer ${options.token}`);
  }

  return headers;
};

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const body = isFormDataBody(options.body)
    ? options.body
    : options.body === undefined
      ? undefined
      : JSON.stringify(options.body);

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: buildHeaders(options),
    body,
  });

  if (response.status === 204) {
    return undefined as T;
  }

  let payload: unknown = null;
  try {
    payload = await response.json();
  } catch {
    payload = null;
  }

  if (!response.ok) {
    const message = (payload as ApiErrorResponse | null)?.error || 'Request failed';
    throw new ApiError(response.status, message);
  }

  return options.rawResponse ? (payload as T) : (payload as ApiResponse<T>).data;
}

export async function apiBlobRequest(path: string, token: string): Promise<Blob> {
  const response = await fetch(apiUrlForPath(path), {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    let message = 'Request failed';
    try {
      const payload = (await response.json()) as ApiErrorResponse;
      message = payload.error || message;
    } catch {
      message = response.statusText || message;
    }
    throw new ApiError(response.status, message);
  }

  return response.blob();
}

const parseServerEventBlock = <T>(block: string): T | null => {
  const data = block
    .split(/\r?\n/)
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice(5).trimStart())
    .join('\n');

  if (!data) {
    return null;
  }

  return JSON.parse(data) as T;
};

export async function apiEventStream<T>(
  path: string,
  token: string,
  signal: AbortSignal,
  onEvent: (event: T) => void,
): Promise<void> {
  const response = await fetch(apiUrlForPath(path), {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'text/event-stream',
    },
    signal,
  });

  if (!response.ok) {
    let message = 'Request failed';
    try {
      const payload = (await response.json()) as ApiErrorResponse;
      message = payload.error || message;
    } catch {
      message = response.statusText || message;
    }
    throw new ApiError(response.status, message);
  }
  if (!response.body) {
    throw new ApiError(response.status, 'Realtime stream is unavailable');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (!signal.aborted) {
      const { value, done } = await reader.read();
      if (done) {
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const blocks = buffer.split(/\r?\n\r?\n/);
      buffer = blocks.pop() || '';

      blocks.forEach((block) => {
        const event = parseServerEventBlock<T>(block);
        if (event) {
          onEvent(event);
        }
      });
    }
  } finally {
    reader.releaseLock();
  }
}

export const listPublicCars = () => apiRequest<Car[]>('/cars?available=true&sort=price_per_day&order=asc');

export const listPublishedNews = () => apiRequest<NewsArticle[]>('/news?sort=published_at&order=desc');

export const getCarRecommendation = (message: string) =>
  apiRequest<CarRecommendationResponse>('/ai/car-recommendation', {
    method: 'POST',
    body: { message },
    rawResponse: true,
  });

export const translateTexts = (texts: string[], targetLang = 'UK') =>
  apiRequest<TranslationResponse>('/translate', {
    method: 'POST',
    body: {
      target_lang: targetLang,
      texts,
    },
    rawResponse: true,
  });

export const login = (payload: LoginPayload) =>
  apiRequest<AuthResponse>('/auth/login', {
    method: 'POST',
    body: payload,
  });

export const registerUser = (payload: RegisterPayload) =>
  apiRequest<AuthResponse>('/auth/register', {
    method: 'POST',
    body: payload,
  });

export const googleLogin = (payload: GoogleLoginPayload) =>
  apiRequest<AuthResponse>('/auth/google', {
    method: 'POST',
    body: payload,
  });

export const getCurrentUser = (token: string) =>
  apiRequest<User>('/auth/me', {
    token,
  });

export const updateCurrentUser = (token: string, payload: UpdateCurrentUserPayload) =>
  apiRequest<User>('/auth/me', {
    method: 'PATCH',
    token,
    body: payload,
  });

export const createRentalOrder = (token: string, payload: RentalOrderInput) =>
  apiRequest<RentalOrder>('/rental-orders', {
    method: 'POST',
    token,
    body: payload,
  });

export const listMyRentalOrders = (token: string) =>
  apiRequest<RentalOrder[]>('/rental-orders', {
    token,
  });

export const listAdminCars = (token: string) =>
  apiRequest<Car[]>('/admin/cars?sort=created_at&order=desc', {
    token,
  });

export const listAdminUsers = (token: string) =>
  apiRequest<User[]>('/admin/users', {
    token,
  });

export const listAdminNews = (token: string) =>
  apiRequest<NewsArticle[]>('/admin/news?sort=created_at&order=desc', {
    token,
  });

export const rateAdminUser = (token: string, id: number, payload: RateUserPayload) =>
  apiRequest<User>(`/admin/users/${id}/rating`, {
    method: 'PATCH',
    token,
    body: payload,
  });

const imageUploadForm = (file: File) => {
  const formData = new FormData();
  formData.append('image', file);
  return formData;
};

export const uploadAdminCarImage = (token: string, file: File) =>
  apiRequest<ImageUploadResponse>('/admin/uploads/car-image', {
    method: 'POST',
    token,
    body: imageUploadForm(file),
  });

export const uploadAdminNewsImage = (token: string, file: File) =>
  apiRequest<ImageUploadResponse>('/admin/uploads/news-image', {
    method: 'POST',
    token,
    body: imageUploadForm(file),
  });

export const createAdminCar = (token: string, payload: CarInput) =>
  apiRequest<Car>('/admin/cars', {
    method: 'POST',
    token,
    body: payload,
  });

export const updateAdminCar = (token: string, id: number, payload: CarInput) =>
  apiRequest<Car>(`/admin/cars/${id}`, {
    method: 'PUT',
    token,
    body: payload,
  });

export const deleteAdminCar = (token: string, id: number) =>
  apiRequest<void>(`/admin/cars/${id}`, {
    method: 'DELETE',
    token,
  });

export const createAdminNews = (token: string, payload: NewsInput) =>
  apiRequest<NewsArticle>('/admin/news', {
    method: 'POST',
    token,
    body: payload,
  });

export const updateAdminNews = (token: string, id: number, payload: NewsInput) =>
  apiRequest<NewsArticle>(`/admin/news/${id}`, {
    method: 'PUT',
    token,
    body: payload,
  });

export const deleteAdminNews = (token: string, id: number) =>
  apiRequest<void>(`/admin/news/${id}`, {
    method: 'DELETE',
    token,
  });

const supportMessageForm = (message: string, files: File[] = []) => {
  const formData = new FormData();
  formData.append('message', message);
  files.forEach((file) => formData.append('files', file));
  return formData;
};

export const getSupportConversation = (token: string) =>
  apiRequest<SupportConversation>('/support/conversation', {
    token,
  });

export const sendSupportMessage = (token: string, message: string, files: File[] = []) =>
  apiRequest<SupportMessage>('/support/messages', {
    method: 'POST',
    token,
    body: supportMessageForm(message, files),
  });

export const listAdminSupportConversations = (token: string) =>
  apiRequest<SupportConversation[]>('/admin/support/conversations', {
    token,
  });

export const getAdminSupportConversation = (token: string, id: number) =>
  apiRequest<SupportConversation>(`/admin/support/conversations/${id}`, {
    token,
  });

export const replyAdminSupportMessage = (token: string, conversationID: number, message: string) =>
  apiRequest<SupportMessage>(`/admin/support/conversations/${conversationID}/messages`, {
    method: 'POST',
    token,
    body: { message },
  });

export const updateAdminSupportConversationStatus = (token: string, conversationID: number, status: 'open' | 'closed') =>
  apiRequest<SupportConversation>(`/admin/support/conversations/${conversationID}/status`, {
    method: 'PATCH',
    token,
    body: { status },
  });

export const downloadSupportAttachment = (token: string, attachment: SupportAttachment) => apiBlobRequest(attachment.file_url, token);

export const streamSupportEvents = (token: string, signal: AbortSignal, onEvent: (event: SupportRealtimeEvent) => void) =>
  apiEventStream<SupportRealtimeEvent>('/support/events', token, signal, onEvent);

export const streamAdminSupportEvents = (token: string, signal: AbortSignal, onEvent: (event: SupportRealtimeEvent) => void) =>
  apiEventStream<SupportRealtimeEvent>('/admin/support/events', token, signal, onEvent);
