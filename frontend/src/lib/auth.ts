// Date: 2026-05-25
// Author: XinYang Li

const AUTH_TOKEN_KEY = "agenthub-token";

/**
 * Reads the current auth token from local storage in the browser.
 * @returns The stored token string or null when the token does not exist.
 */
export function getStoredToken(): string | null {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem(AUTH_TOKEN_KEY);
}

/**
 * Persists the auth token to local storage in the browser.
 * @param token The token string returned by the backend login or register endpoint.
 */
export function setStoredToken(token: string): void {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(AUTH_TOKEN_KEY, token);
}

/**
 * Removes the auth token from local storage.
 */
export function clearStoredToken(): void {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(AUTH_TOKEN_KEY);
}
