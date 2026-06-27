export interface ApiSuccessEnvelope<T> {
  data: T;
}

export interface ApiErrorBody {
  code: string;
  message: string;
  fields?: Record<string, string[]>;
}

export interface ApiErrorEnvelope {
  error: ApiErrorBody;
}

export type ApiEnvelope<T> = ApiSuccessEnvelope<T> | ApiErrorEnvelope;

export function isApiErrorEnvelope<T>(e: ApiEnvelope<T>): e is ApiErrorEnvelope {
  return typeof e === "object" && e !== null && "error" in e;
}

export interface RequestOptions extends RequestInit {
  idempotencyKey?: string;
  timeout?: number;
}
