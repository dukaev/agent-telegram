import React, {FormEvent, useState} from "react";
import {Alert, Button, Input, Spinner} from "@heroui/react";

import {postAuthState} from "../api";
import {useCountdown} from "../hooks";
import {AuthState} from "../types";
import {CheckIcon, LockIcon} from "./Brand";

type ScreenProps = {
  state: AuthState;
  onState: (state: AuthState) => void;
};

function ErrorAlert({message}: {message: string}) {
  if (!message) {
    return null;
  }
  return (
    <Alert className="state-alert" status="danger" role="alert">
      <Alert.Content>
        <Alert.Title>Could not continue</Alert.Title>
        <Alert.Description>{message}</Alert.Description>
      </Alert.Content>
    </Alert>
  );
}

function useAuthAction(onState: (state: AuthState) => void) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const run = async (path: string, payload: unknown) => {
    setBusy(true);
    setError("");
    try {
      const result = await postAuthState(path, payload);
      onState(result.state);
      if (!result.ok) {
        setError(result.state.error || "Could not continue.");
        return false;
      }
      return true;
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not continue.");
      return false;
    } finally {
      setBusy(false);
    }
  };

  return {busy, error, run};
}

function formatCountdown(seconds: number | null) {
  if (seconds === null) {
    return "The code refreshes automatically";
  }
  if (seconds <= 0) {
    return "Refreshing the QR code…";
  }
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  return `Code refreshes in ${minutes}:${rest.toString().padStart(2, "0")}`;
}

function QRScreen({state, onState}: ScreenProps) {
  const countdown = useCountdown(state.expires);
  const {busy, error, run} = useAuthAction(onState);

  return (
    <section className="auth-screen qr-screen">
      <ol className="qr-instructions" aria-label="How to scan the QR code">
        <li><span>1</span><p>Open Telegram on your phone</p></li>
        <li><span>2</span><p>Go to <strong>Settings → Devices</strong></p></li>
        <li><span>3</span><p>Tap <strong>Link Desktop Device</strong> and point the camera at the code</p></li>
      </ol>

      <div className="qr-stage" aria-live="polite" aria-busy={!state.qrImage}>
        {state.qrImage ? (
          <img
            alt="QR code for signing in to Telegram"
            className="qr-image"
            key={state.expires || state.qrImage.slice(-24)}
            src={state.qrImage}
          />
        ) : (
          <div className="qr-placeholder">
            <Spinner size="lg" />
            <span>Creating a secure QR code…</span>
          </div>
        )}
      </div>
      <p className="qr-expiry" aria-live="polite">{formatCountdown(countdown)}</p>

      <ErrorAlert message={error || state.error || ""} />

      {state.mock?.enabled && (
        <div className="mock-panel">
          <div>
            <strong>Mock mode</strong>
            <span>Continue to access setup without Telegram.</span>
          </div>
          <Button isDisabled={busy} size="sm" type="button" onClick={() => void run("/auth/mock/advance", {action: "qr_scan"})}>
            Simulate QR scan
          </Button>
        </div>
      )}

      <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "phone"})}>
        Sign in with phone number
      </button>
      <div className="trust-note">
        <LockIcon />
        <span>This page is only available on this computer and closes after sign-in.</span>
      </div>
    </section>
  );
}

function PhoneScreen({state, onState}: ScreenProps) {
  const [phone, setPhone] = useState("");
  const {busy, error, run} = useAuthAction(onState);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    void run("/auth/mode", {mode: "code", phone});
  };

  return (
    <form className="auth-screen auth-form" onSubmit={submit}>
      <ErrorAlert message={error || state.error || ""} />
      <label className="field-label" htmlFor="phone-number">
        Phone number
        <Input
          autoComplete="tel"
          autoFocus
          fullWidth
          id="phone-number"
          inputMode="tel"
          placeholder="+90 555 123 45 67"
          required
          type="tel"
          value={phone}
          onChange={(event) => setPhone(event.currentTarget.value)}
        />
      </label>
      <p className="field-help">Use international format with a country code. Telegram will send a code to the app.</p>
      <Button className="primary-action" isDisabled={busy || phone.trim().length < 7} type="submit">
        {busy ? "Sending code…" : "Get code"}
      </Button>
      <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "qr"})}>
        Back to QR code
      </button>
      <div className="trust-note">
        <LockIcon />
        <span>Your number is only used to sign in to Telegram and is not stored after authentication.</span>
      </div>
    </form>
  );
}

function CodeScreen({state, onState}: ScreenProps) {
  const [code, setCode] = useState("");
  const {busy, error, run} = useAuthAction(onState);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    void run("/auth/verify", {code});
  };

  return (
    <form className="auth-screen auth-form" onSubmit={submit}>
      <ErrorAlert message={error || state.error || ""} />
      <label className="field-label" htmlFor="telegram-code">
        Telegram code
        <Input
          aria-describedby="code-help"
          autoComplete="one-time-code"
          autoFocus
          fullWidth
          id="telegram-code"
          inputMode="numeric"
          maxLength={8}
          placeholder="•••••"
          required
          value={code}
          onChange={(event) => setCode(event.currentTarget.value.replace(/\s/g, ""))}
        />
      </label>
      <p className="field-help" id="code-help">The code was sent for {state.phone}. Never share it with anyone.</p>

      {state.mock?.code && (
        <div className="mock-panel">
          <div><strong>Mock code</strong><span>{state.mock.code}</span></div>
          <Button isDisabled={busy} size="sm" type="button" variant="secondary" onClick={() => void run("/auth/verify", {code: state.mock?.code ?? ""})}>
            Use code
          </Button>
        </div>
      )}

      <Button className="primary-action" isDisabled={busy || code.length < 4} type="submit">
        {busy ? "Verifying…" : "Continue"}
      </Button>
      <div className="secondary-actions">
        <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "phone"})}>
          Change number
        </button>
        <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "qr"})}>
          Sign in with QR code
        </button>
      </div>
    </form>
  );
}

function PasswordScreen({state, onState}: ScreenProps) {
  const [password, setPassword] = useState("");
  const {busy, error, run} = useAuthAction(onState);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    void run("/auth/password", {password});
  };

  return (
    <form className="auth-screen auth-form" onSubmit={submit}>
      <ErrorAlert message={error || state.error || ""} />
      <label className="field-label" htmlFor="telegram-password">
        Cloud password
        <Input
          aria-describedby="password-help"
          autoComplete="current-password"
          autoFocus
          fullWidth
          id="telegram-password"
          required
          type="password"
          value={password}
          onChange={(event) => setPassword(event.currentTarget.value)}
        />
      </label>
      <p className="field-help" id="password-help">{state.hint || "Enter your Telegram two-step verification password."}</p>

      {state.mock?.password && (
        <div className="mock-panel">
          <div><strong>Mock password</strong><span>{state.mock.password}</span></div>
          <Button isDisabled={busy} size="sm" type="button" variant="secondary" onClick={() => void run("/auth/password", {password: state.mock?.password ?? ""})}>
            Use password
          </Button>
        </div>
      )}

      <Button className="primary-action" isDisabled={busy || !password} type="submit">
        {busy ? "Verifying…" : "Sign in"}
      </Button>
      <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "phone"})}>
        Start over
      </button>
    </form>
  );
}

function DoneScreen() {
  return (
    <section className="auth-screen done-screen" aria-live="polite">
      <div className="success-mark"><CheckIcon /></div>
      <div>
        <h2>Authentication complete</h2>
        <p>Your session has been saved. You can close this page and return to the terminal.</p>
      </div>
    </section>
  );
}

export function AuthScreen(props: ScreenProps) {
  switch (props.state.mode) {
    case "qr":
      return <QRScreen {...props} />;
    case "phone":
      return <PhoneScreen {...props} />;
    case "code":
      return <CodeScreen {...props} />;
    case "password":
      return <PasswordScreen {...props} />;
    case "done":
      return <DoneScreen />;
    default:
      return null;
  }
}
