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
        <Alert.Title>Не удалось продолжить</Alert.Title>
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
        setError(result.state.error || "Не удалось продолжить.");
        return false;
      }
      return true;
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Не удалось продолжить.");
      return false;
    } finally {
      setBusy(false);
    }
  };

  return {busy, error, run};
}

function formatCountdown(seconds: number | null) {
  if (seconds === null) {
    return "Код обновится автоматически";
  }
  if (seconds <= 0) {
    return "Обновляю QR-код…";
  }
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  return `Код обновится через ${minutes}:${rest.toString().padStart(2, "0")}`;
}

function QRScreen({state, onState}: ScreenProps) {
  const countdown = useCountdown(state.expires);
  const {busy, error, run} = useAuthAction(onState);

  return (
    <section className="auth-screen qr-screen">
      <ol className="qr-instructions" aria-label="Как отсканировать QR-код">
        <li><span>1</span><p>Открой Telegram на телефоне</p></li>
        <li><span>2</span><p>Перейди в <strong>Настройки → Устройства</strong></p></li>
        <li><span>3</span><p>Нажми <strong>Подключить устройство</strong> и наведи камеру на код</p></li>
      </ol>

      <div className="qr-stage" aria-live="polite" aria-busy={!state.qrImage}>
        {state.qrImage ? (
          <img
            alt="QR-код для входа в Telegram"
            className="qr-image"
            key={state.expires || state.qrImage.slice(-24)}
            src={state.qrImage}
          />
        ) : (
          <div className="qr-placeholder">
            <Spinner size="lg" />
            <span>Создаю защищённый QR-код…</span>
          </div>
        )}
      </div>
      <p className="qr-expiry" aria-live="polite">{formatCountdown(countdown)}</p>

      <ErrorAlert message={error || state.error || ""} />

      {state.mock?.enabled && (
        <div className="mock-panel">
          <div>
            <strong>Тестовый режим</strong>
            <span>Можно перейти к настройке доступа без Telegram.</span>
          </div>
          <Button isDisabled={busy} size="sm" type="button" onClick={() => void run("/auth/mock/advance", {action: "qr_scan"})}>
            Имитировать сканирование
          </Button>
        </div>
      )}

      <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "phone"})}>
        Войти по номеру телефона
      </button>
      <div className="trust-note">
        <LockIcon />
        <span>Страница доступна только на этом компьютере и закроется после завершения входа.</span>
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
        Номер телефона
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
      <p className="field-help">Используй международный формат с кодом страны. Telegram отправит код в приложение.</p>
      <Button className="primary-action" isDisabled={busy || phone.trim().length < 7} type="submit">
        {busy ? "Отправляю код…" : "Получить код"}
      </Button>
      <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "qr"})}>
        Вернуться к QR-коду
      </button>
      <div className="trust-note">
        <LockIcon />
        <span>Номер используется только для входа в Telegram и не сохраняется после авторизации.</span>
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
        Код из Telegram
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
      <p className="field-help" id="code-help">Код отправлен для {state.phone}. Не сообщай его другим людям.</p>

      {state.mock?.code && (
        <div className="mock-panel">
          <div><strong>Тестовый код</strong><span>{state.mock.code}</span></div>
          <Button isDisabled={busy} size="sm" type="button" variant="secondary" onClick={() => void run("/auth/verify", {code: state.mock?.code ?? ""})}>
            Использовать
          </Button>
        </div>
      )}

      <Button className="primary-action" isDisabled={busy || code.length < 4} type="submit">
        {busy ? "Проверяю…" : "Продолжить"}
      </Button>
      <div className="secondary-actions">
        <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "phone"})}>
          Изменить номер
        </button>
        <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "qr"})}>
          Войти по QR-коду
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
        Облачный пароль
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
      <p className="field-help" id="password-help">{state.hint || "Введи пароль двухэтапной аутентификации Telegram."}</p>

      {state.mock?.password && (
        <div className="mock-panel">
          <div><strong>Тестовый пароль</strong><span>{state.mock.password}</span></div>
          <Button isDisabled={busy} size="sm" type="button" variant="secondary" onClick={() => void run("/auth/password", {password: state.mock?.password ?? ""})}>
            Использовать
          </Button>
        </div>
      )}

      <Button className="primary-action" isDisabled={busy || !password} type="submit">
        {busy ? "Проверяю…" : "Войти"}
      </Button>
      <button className="text-action" disabled={busy} type="button" onClick={() => void run("/auth/mode", {mode: "phone"})}>
        Начать заново
      </button>
    </form>
  );
}

function DoneScreen() {
  return (
    <section className="auth-screen done-screen" aria-live="polite">
      <div className="success-mark"><CheckIcon /></div>
      <div>
        <h2>Авторизация завершена</h2>
        <p>Сессия сохранена. Эту страницу можно закрыть и вернуться в терминал.</p>
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
