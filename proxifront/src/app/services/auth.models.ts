export interface RegisterRequest {
  display_name: string;
  email: string;
  password: string;
}

export interface RegisterResponse {
  id: string;
  display_name: string;
  email: string;
}

export interface ApiErrorResponse {
  error: string;
  message: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthenticatedUser {
  id: string;
  display_name: string;
  email: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: AuthenticatedUser;
}

export interface RefreshRequest {
  refresh_token: string;
}

export type RefreshResponse = LoginResponse;
