export interface ApiErrorResponse {
  message: string;
  data: null;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
}