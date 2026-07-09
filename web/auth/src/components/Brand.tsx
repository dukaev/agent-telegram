import React from "react";

import {AuthMode} from "../types";

export function TelegramMark() {
  return (
    <span className="telegram-mark" aria-hidden="true">
      <svg viewBox="0 0 32 32" role="img">
        <path d="M25.9 7.1 22.5 24c-.3 1.2-1 1.5-2 1l-5.2-3.8-2.5 2.4c-.3.3-.5.5-1 .5l.4-5.3 9.6-8.7c.4-.4-.1-.6-.6-.2L9.3 17.4l-5.1-1.6c-1.1-.3-1.1-1.1.2-1.6L24.3 6.5c.9-.3 1.7.2 1.6.6Z" />
      </svg>
    </span>
  );
}

function stepForMode(mode?: AuthMode) {
  if (mode === "setup") {
    return 2;
  }
  if (mode === "done") {
    return 2;
  }
  return 1;
}

export function AuthHeader({mode, title, description}: {mode?: AuthMode; title: string; description: string}) {
  const step = stepForMode(mode);
  const complete = mode === "done";
  return (
    <header className="auth-header">
      <div className="brand-row">
        <div className="brand-lockup">
          <TelegramMark />
          <span>Agent Telegram</span>
        </div>
        <div className="step-label" aria-label={complete ? "Авторизация завершена" : `Шаг ${step} из 2`}>
          {complete ? "Готово" : `Шаг ${step} из 2`}
        </div>
      </div>
      <div className="step-track" aria-hidden="true">
        <span className="is-active" />
        <span className={step === 2 ? "is-active" : ""} />
      </div>
      <div className="heading-copy">
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
    </header>
  );
}

export function LockIcon() {
  return (
    <svg className="inline-icon" viewBox="0 0 20 20" aria-hidden="true">
      <path d="M6.5 8V6.5a3.5 3.5 0 1 1 7 0V8m-8 0h9A1.5 1.5 0 0 1 16 9.5v6A1.5 1.5 0 0 1 14.5 17h-9A1.5 1.5 0 0 1 4 15.5v-6A1.5 1.5 0 0 1 5.5 8Z" />
    </svg>
  );
}

export function CheckIcon() {
  return (
    <svg viewBox="0 0 28 28" aria-hidden="true">
      <path d="m7 14.5 4.4 4.4L21.5 8.8" />
    </svg>
  );
}
