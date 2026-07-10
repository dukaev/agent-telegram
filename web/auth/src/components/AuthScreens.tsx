import React, {useState} from "react";
import {Alert, Button, Spinner} from "@heroui/react";

import {postAuthState} from "../api";
import {useCountdown} from "../hooks";
import {AuthSession, AuthState} from "../types";
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

function sessionProviderLabel(provider: string) {
  switch (provider) {
    case "keychain":
      return "macOS Keychain";
    case "memory":
      return "temporary memory";
    default:
      return provider;
  }
}

function SessionSaveNote({session}: {session?: AuthSession}) {
  if (!session?.saveByDefault) {
    return null;
  }
  const provider = sessionProviderLabel(session.provider);
  return (
    <div className="session-save-note">
      <LockIcon />
      <span>
        Your session will be saved to <strong>{provider}</strong> under profile <strong>{session.profile}</strong>.
      </span>
    </div>
  );
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

      {state.session?.error && !state.error && !error && (
        <Alert className="state-alert" status="warning" role="status">
          <Alert.Content>
            <Alert.Title>Session storage needs attention</Alert.Title>
            <Alert.Description>{state.session.error}</Alert.Description>
          </Alert.Content>
        </Alert>
      )}

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

      <SessionSaveNote session={state.session} />
      <div className="trust-note">
        <LockIcon />
        <span>This page is only available on this computer and closes after sign-in.</span>
      </div>
    </section>
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
    case "done":
      return <DoneScreen />;
    default:
      return null;
  }
}
