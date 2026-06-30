import { ApiRequestError } from "./apiRequest";

export function apiErrorMessage(error: unknown, fallback: string) {
  return error instanceof ApiRequestError ? error.message : fallback;
}
