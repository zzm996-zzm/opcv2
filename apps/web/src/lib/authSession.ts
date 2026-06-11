import { useSyncExternalStore } from "react";

import type { LoginResponse, User } from "./authApi";

type AuthState = {
  accessToken: string | null;
  expiresAt: string | null;
  ready: boolean;
  user: User | null;
};

let state: AuthState = {
  accessToken: null,
  expiresAt: null,
  ready: false,
  user: null
};

const listeners = new Set<() => void>();

function notify() {
  listeners.forEach((listener) => listener());
}

export const authSession = {
  get() {
    return state;
  },

  subscribe(listener: () => void) {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },

  set(result: LoginResponse) {
    state = {
      accessToken: result.access_token,
      expiresAt: result.access_token_expires_at,
      ready: true,
      user: result.user
    };
    notify();
  },

  finishRestore() {
    if (state.ready) return;
    state = { ...state, ready: true };
    notify();
  },

  clear() {
    if (!state.accessToken && !state.user && state.ready) return;
    state = { accessToken: null, expiresAt: null, ready: true, user: null };
    notify();
  }
};

export function useAuthSession() {
  return useSyncExternalStore(authSession.subscribe, authSession.get, authSession.get);
}
