import { API_CONFIG } from "./config";
import { ApiError } from "./error";
import { ApiEnvelope, isApiErrorEnvelope, RequestOptions } from "./types";

async function request<T>(method: string, path: string, body?: unknown, options?: RequestOptions): Promise<T> {
  const url = `${API_CONFIG.baseURL}${path}`;
  
  const headers = new Headers(options?.headers);
  if (body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (options?.idempotencyKey) {
    headers.set("Idempotency-Key", options.idempotencyKey);
  }

  let response: Response;
  try {
    const fetchOptions: RequestInit = {
      ...options,
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    };
    
    response = await fetch(url, fetchOptions);
  } catch (error) {
    throw ApiError.fromNetwork(error);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  let envelope: ApiEnvelope<T>;
  try {
    envelope = await response.json();
  } catch {
    throw new ApiError("Failed to parse JSON", "INVALID_RESPONSE", response.status);
  }

  if (!response.ok) {
    if (isApiErrorEnvelope(envelope)) {
      throw ApiError.fromEnvelope(envelope.error, response.status);
    }
    throw new ApiError("Unknown server error", "UNKNOWN_ERROR", response.status);
  }

  if (isApiErrorEnvelope(envelope)) {
    throw new ApiError("Success response contained error body", "CONTRACT_VIOLATION", response.status);
  }

  return envelope.data;
}

export const apiClient = {
  get<T>(path: string, options?: RequestOptions): Promise<T> {
    return request<T>("GET", path, undefined, options);
  },
  post<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return request<T>("POST", path, body, options);
  },
  patch<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return request<T>("PATCH", path, body, options);
  },
  put<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return request<T>("PUT", path, body, options);
  },
  delete<T>(path: string, options?: RequestOptions): Promise<T> {
    return request<T>("DELETE", path, undefined, options);
  },
} as const;

export type ApiClient = typeof apiClient;
