import { ApiErrorBody } from "./types";

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;
  readonly fields?: Record<string, string[]>;

  constructor(message: string, code: string, status: number, fields?: Record<string, string[]>) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.fields = fields;
  }

  get isNetworkError() {
    return this.status === 0;
  }

  get hasFieldErrors() {
    return !!this.fields && Object.keys(this.fields).length > 0;
  }

  get messageKey() {
    if (this.isNetworkError) return "network_error";
    return this.code || "unknown_error";
  }

  static fromEnvelope(body: ApiErrorBody, status: number) {
    return new ApiError(body.message, body.code, status, body.fields);
  }

  static fromNetwork(cause: unknown) {
    return new ApiError(cause instanceof Error ? cause.message : "Network error", "NETWORK_ERROR", 0);
  }
}
