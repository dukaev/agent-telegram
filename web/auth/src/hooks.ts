import {useCallback, useEffect, useState} from "react";

import {fetchAuthState} from "./api";
import {AuthState} from "./types";

export function useAuthState() {
  const [state, setState] = useState<AuthState | null>(null);
  const [loadError, setLoadError] = useState("");

  const load = useCallback(async () => {
    try {
      const result = await fetchAuthState();
      if (!result.ok) {
        setLoadError(result.state.error || "Could not load the authentication state.");
        return;
      }
      setState(result.state);
      setLoadError("");
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Could not load the authentication state.");
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!state || state.mode === "done") {
      return undefined;
    }
    const delay = Math.max(1000, (state.refresh ?? 2) * 1000);
    const timer = window.setTimeout(() => void load(), delay);
    return () => window.clearTimeout(timer);
  }, [load, state]);

  return {state, setState, loadError, retry: load};
}

export function useCountdown(expires?: string) {
  const calculate = useCallback(() => {
    if (!expires) {
      return null;
    }
    const target = Date.parse(expires);
    if (!Number.isFinite(target)) {
      return null;
    }
    return Math.max(0, Math.ceil((target - Date.now()) / 1000));
  }, [expires]);
  const [seconds, setSeconds] = useState<number | null>(calculate);

  useEffect(() => {
    setSeconds(calculate());
    if (!expires) {
      return undefined;
    }
    const timer = window.setInterval(() => setSeconds(calculate()), 1000);
    return () => window.clearInterval(timer);
  }, [calculate, expires]);

  return seconds;
}
