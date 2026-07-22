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
