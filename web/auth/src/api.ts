import {AuthState, defaultPolicy, PeersState} from "./types";

const token = new URLSearchParams(window.location.search).get("t") ?? "";

function apiURL(path: string) {
  const suffix = token ? `?t=${encodeURIComponent(token)}` : "";
  return `${path}${suffix}`;
}

async function readJSON<T>(response: Response): Promise<T> {
  return (await response.json()) as T;
}

function normalizeState(state: AuthState): AuthState {
  return {
    ...state,
    policy: state.policy ?? structuredClone(defaultPolicy),
  };
}

export async function fetchAuthState() {
  const response = await fetch(apiURL("/auth/state"), {
    headers: {Accept: "application/json"},
  });
  const state = normalizeState(await readJSON<AuthState>(response));
  return {ok: response.ok, state};
}

export async function postAuthState(path: string, payload: unknown) {
  const response = await fetch(apiURL(path), {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
  const state = normalizeState(await readJSON<AuthState>(response));
  return {ok: response.ok, state};
}

export async function fetchPeers() {
  const response = await fetch(apiURL("/auth/peers"), {
    headers: {Accept: "application/json"},
  });
  const state = await readJSON<PeersState>(response);
  state.peers ??= [];
  return {ok: response.ok, state};
}
