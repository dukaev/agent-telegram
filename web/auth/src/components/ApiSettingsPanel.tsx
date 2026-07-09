import React, {FormEvent, useEffect, useState} from "react";
import {Alert, Button, Chip, Input} from "@heroui/react";

import {postAuthState} from "../api";
import {AuthAPI, AuthState} from "../types";

export function ApiSettingsPanel({api, onUpdated}: {api: AuthAPI; onUpdated: (state: AuthState) => void}) {
  const [open, setOpen] = useState(false);
  const [appId, setAppId] = useState(api.appId ? String(api.appId) : "");
  const [appHash, setAppHash] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setAppId(api.appId ? String(api.appId) : "");
  }, [api.appId]);

  const save = async (payload: {appId?: string; appHash?: string; useDefault?: boolean}) => {
    setSaving(true);
    setError("");
    try {
      const result = await postAuthState("/auth/api", payload);
      onUpdated(result.state);
      if (!result.ok) {
        setError(result.state.error || "Не удалось обновить настройки Telegram API.");
        return;
      }
      setAppHash("");
      setOpen(false);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Не удалось обновить настройки Telegram API.");
    } finally {
      setSaving(false);
    }
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    void save({appId, appHash});
  };

  return (
    <section className="advanced-panel">
      <button
        aria-expanded={open}
        className="advanced-trigger"
        disabled={!api.canEdit || saving}
        type="button"
        onClick={() => setOpen((current) => !current)}
      >
        <span>
          <span className="advanced-title">Расширенные настройки</span>
          <span className="advanced-subtitle">Telegram API ID и Hash</span>
        </span>
        <span className="advanced-meta">
          <Chip color={api.default ? "default" : "accent"} size="sm" variant="soft">
            {api.default ? "Стандартный API" : "Свой API"}
          </Chip>
          <span className={`disclosure-icon ${open ? "is-open" : ""}`} aria-hidden="true">⌄</span>
        </span>
      </button>

      {open && (
        <form className="advanced-content" onSubmit={submit}>
          <p className="field-help">
            Большинству пользователей менять эти параметры не нужно. Собственные значения можно получить на my.telegram.org.
          </p>
          {error && (
            <Alert status="danger" role="alert">
              <Alert.Content>
                <Alert.Title>Не удалось сохранить API</Alert.Title>
                <Alert.Description>{error}</Alert.Description>
              </Alert.Content>
            </Alert>
          )}
          <div className="api-fields">
            <label className="field-label" htmlFor="telegram-app-id">
              App ID
              <Input
                fullWidth
                id="telegram-app-id"
                inputMode="numeric"
                required
                value={appId}
                onChange={(event) => setAppId(event.currentTarget.value)}
              />
            </label>
            <label className="field-label" htmlFor="telegram-app-hash">
              App Hash
              <Input
                autoComplete="off"
                fullWidth
                id="telegram-app-hash"
                required
                type="password"
                value={appHash}
                onChange={(event) => setAppHash(event.currentTarget.value)}
              />
            </label>
          </div>
          <div className="api-actions">
            <Button isDisabled={saving} size="sm" type="submit">
              {saving ? "Сохраняю…" : "Применить"}
            </Button>
            {!api.default && (
              <Button isDisabled={saving} size="sm" type="button" variant="secondary" onClick={() => void save({useDefault: true})}>
                Вернуть стандартные
              </Button>
            )}
          </div>
        </form>
      )}
    </section>
  );
}
