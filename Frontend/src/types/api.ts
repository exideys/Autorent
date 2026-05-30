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
}

export interface CarRecommendationResponse {
  answer: string;
  cars: RecommendedCar[];
  total_matches: number;
}
