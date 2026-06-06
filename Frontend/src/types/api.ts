export interface ApiResponse<T> {
  data: T;
}

export interface ApiErrorResponse {
  error?: string;
}

export interface User {
  id: number;
  first_name: string;
  last_name: string;
  name: string;
  email: string;
  rating: number;
  rating_count: number;
  role: 'user' | 'admin' | string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface RegisterPayload {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface GoogleLoginPayload {
  credential: string;
}

export interface UpdateCurrentUserPayload {
  first_name: string;
  last_name: string;
  email: string;
}

export interface RateUserPayload {
  rating: number;
}

export interface CarImage {
  id: number;
  car_id: number;
  image_url: string;
  is_main: boolean;
  sort_order: number;
}

export interface Car {
  id: number;
  brand: string;
  model: string;
  year: number;
  car_class: string;
  body_type: string;
  transmission: string;
  fuel_type: string;
  seats: number;
  doors: number;
  engine_volume?: number;
  horsepower?: number;
  price_per_day: number;
  deposit: number;
  color?: string;
  status: string;
  created_at: string;
  images: CarImage[];
}

export interface CarImageInput {
  image_url: string;
  is_main: boolean;
  sort_order: number;
}

export interface CarInput {
  brand: string;
  model: string;
  year: number;
  car_class: string;
  body_type: string;
  transmission: string;
  fuel_type: string;
  seats: number;
  doors: number;
  engine_volume?: number;
  horsepower?: number;
  price_per_day: number;
  deposit: number;
  color?: string;
  status?: string;
  images?: CarImageInput[];
}

export interface RecommendedCar {
  id: number;
  brand: string;
  model: string;
  year: number;
  car_class: string;
  body_type: string;
  transmission: string;
  fuel_type: string;
  seats: number;
  doors: number;
  engine_volume?: number;
  horsepower: number;
  price_per_day: number;
  deposit: number;
  color: string;
  status: string;
  created_at?: string;
  images?: CarImage[];
}

export interface CarRecommendationResponse {
  answer: string;
  cars: RecommendedCar[];
  total_matches: number;
}

export interface TranslationResponse {
  translations: string[];
}

export interface ImageUploadResponse {
  image_url: string;
  file_id: string;
  drive_url?: string;
}

export interface SupportAttachment {
  id: number;
  message_id: number;
  file_name: string;
  content_type: string;
  file_size: number;
  drive_file_id: string;
  file_url: string;
  drive_url?: string;
  created_at: string;
}

export interface SupportMessage {
  id: number;
  conversation_id: number;
  sender_id: number;
  sender_role: 'user' | 'admin' | string;
  body: string;
  created_at: string;
  attachments?: SupportAttachment[];
}

export interface SupportConversation {
  id: number;
  user_id: number;
  user?: User;
  status: string;
  last_message_at?: string;
  created_at?: string;
  updated_at?: string;
  messages?: SupportMessage[];
}

export interface SupportRealtimeEvent {
  event_type?: 'message' | 'presence' | string;
  conversation_id: number;
  message_id: number;
  user_id: number;
  sender_role: 'user' | 'admin' | string;
  admin_online?: boolean;
}

export interface RentalOrderCarSummary {
  id: number;
  brand: string;
  model: string;
  year: number;
  price_per_day: number;
  deposit: number;
  status: string;
  image_url?: string;
}

export interface RentalOrder {
  id: number;
  user_id: number;
  car_id: number;
  start_date: string;
  end_date: string;
  pickup_location: string;
  pickup_time: string;
  phone: string;
  notes: string;
  total_price: number;
  deposit: number;
  status: string;
  created_at: string;
  updated_at: string;
  car: RentalOrderCarSummary;
}

export interface RentalOrderInput {
  car_id: number;
  start_date: string;
  end_date: string;
  pickup_location: string;
  pickup_time: string;
  phone: string;
  notes?: string;
}

export type NewsStatus = 'draft' | 'published';

export interface NewsArticle {
  id: number;
  title: string;
  summary: string;
  content: string;
  image_url?: string;
  status: NewsStatus | string;
  published_at?: string;
  created_at: string;
  updated_at: string;
}

export interface NewsInput {
  title: string;
  summary: string;
  content: string;
  image_url?: string;
  status: NewsStatus;
}
