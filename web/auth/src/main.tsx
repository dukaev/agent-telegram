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
      return "Подключи этот компьютер к своему аккаунту с помощью приложения Telegram.";
    case "phone":
      return "Получим одноразовый код в Telegram и продолжим настройку.";
    case "code":
      return state.hint || "Введи одноразовый код из Telegram.";
    case "password":
      return "Аккаунт защищён дополнительным облачным паролем.";
    case "setup":
      return "Реши, с какими диалогами сможет взаимодействовать локальный агент.";
    case "done":
      return "Telegram подключён, а локальные разрешения сохранены.";
    default:
      return "Готовлю безопасную локальную страницу авторизации.";
  }
}

function titleFor(mode?: AuthMode, fallback?: string) {
  if (!mode) {
    return "Telegram авторизация";
  }
  return fallback || "Telegram авторизация";
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
                <Alert.Title>Страница потеряла соединение</Alert.Title>
                <Alert.Description>{loadError}</Alert.Description>
              </Alert.Content>
              <Button size="sm" type="button" variant="secondary" onClick={() => void retry()}>
                Повторить
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
              <span>Подготавливаю авторизацию…</span>
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
