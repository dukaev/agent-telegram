import React from "react";
import {createRoot} from "react-dom/client";
import {Alert, Button, Card, Spinner} from "@heroui/react";

import "./styles.css";

import {AccessSetup} from "./components/AccessSetup";
import {ApiSettingsPanel} from "./components/ApiSettingsPanel";
import {AuthHeader} from "./components/Brand";
import {AuthScreen} from "./components/AuthScreens";
import {useAuthState} from "./hooks";
import {AuthMode, AuthState} from "./types";

function descriptionFor(state?: AuthState) {
  switch (state?.mode) {
    case "qr":
      return "Connect this computer to your account using the Telegram app.";
    case "phone":
      return "Get a one-time code in Telegram, then continue setup.";
    case "code":
      return state.hint || "Enter the one-time code from Telegram.";
    case "password":
      return "Your account is protected by an additional cloud password.";
    case "setup":
      return "Choose which chats the local agent can interact with.";
    case "done":
      return "Telegram is connected and your local permissions are saved.";
    default:
      return "Preparing a secure local authentication page.";
  }
}

function titleFor(mode?: AuthMode, fallback?: string) {
  if (!mode) {
    return "Telegram authentication";
  }
  return fallback || "Telegram authentication";
}

function App() {
  const {state, setState, loadError, retry} = useAuthState();
  const setupMode = state?.mode === "setup";

  return (
    <main className="auth-shell">
      <Card className={`auth-card ${setupMode ? "auth-card--setup" : ""}`} variant="default">
        <Card.Header className="card-header">
          <AuthHeader
            description={descriptionFor(state ?? undefined)}
            mode={state?.mode}
            title={titleFor(state?.mode, state?.title)}
          />
        </Card.Header>

        <Card.Content className="auth-content">
          {loadError && (
            <Alert status="danger" role="alert">
              <Alert.Content>
                <Alert.Title>Connection lost</Alert.Title>
                <Alert.Description>{loadError}</Alert.Description>
              </Alert.Content>
              <Button size="sm" type="button" variant="secondary" onClick={() => void retry()}>
                Retry
              </Button>
            </Alert>
          )}

          {state ? (
            state.mode === "setup" ? (
              <AccessSetup state={state} onState={setState} />
            ) : (
              <>
                <AuthScreen state={state} onState={setState} />
                {(state.mode === "qr" || state.mode === "phone") && state.api.canEdit && (
                  <ApiSettingsPanel api={state.api} onUpdated={setState} />
                )}
              </>
            )
          ) : (
            <div className="page-loading" aria-live="polite">
              <Spinner size="lg" />
              <span>Preparing authentication…</span>
            </div>
          )}
        </Card.Content>
      </Card>
    </main>
  );
}

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
