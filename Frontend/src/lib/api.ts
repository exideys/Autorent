import type {
  ApiErrorResponse,
  ApiResponse,
  AuthResponse,
  Car,
  CarInput,
  CarRecommendationResponse,
  ImageUploadResponse,
  LoginPayload,
  NewsArticle,
  NewsInput,
  RateUserPayload,
  RegisterPayload,
  RentalOrder,
  RentalOrderInput,
  TranslationResponse,
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

export const getCurrentUser = (token: string) =>
  apiRequest<User>('/auth/me', {
    token,
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
