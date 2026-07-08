import React, { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  Alert,
  Button,
  Card,
  Chip,
  Input,
  Spinner,
  Switch,
  TextArea,
} from "@heroui/react";

import "./styles.css";

type Safeties = {
  read: boolean;
  write: boolean;
  destructive: boolean;
  paid: boolean;
};

type PeerTypes = {
  users: boolean;
  groups: boolean;
  channels: boolean;
  bots: boolean;
};

type Policy = {
  version: number;
  safeties: Safeties;
  peerTypes: PeerTypes;
  allowPeers?: string[];
  denyPeers?: string[];
};

type AuthState = {
  title: string;
  message?: string;
  error?: string;
  mode: "qr" | "code" | "password" | "setup" | "done";
  completed: boolean;
  phone?: string;
  hint?: string;
  qrImage?: string;
  qrLink?: string;
  expires?: string;
  refresh?: number;
  api: AuthAPI;
  policy: Policy;
};

type AuthAPI = {
  appId: number;
  default: boolean;
  canEdit: boolean;
};

type PeerOption = {
  peer: string;
  title: string;
  username?: string;
  type: "user" | "group" | "channel" | "bot" | "";
  id?: number;
};

type PeersState = {
  peers: PeerOption[];
  count: number;
  loaded: boolean;
  loading: boolean;
  error?: string;
};

const emptyPolicy: Policy = {
  version: 1,
  safeties: {
    read: true,
    write: true,
    destructive: false,
    paid: false,
  },
  peerTypes: {
    users: true,
    groups: true,
    channels: true,
    bots: true,
  },
  allowPeers: [],
  denyPeers: [],
};

const token = new URLSearchParams(window.location.search).get("t") ?? "";

function apiURL(path: string) {
  const suffix = token ? `?t=${encodeURIComponent(token)}` : "";
  return `${path}${suffix}`;
}

async function readState(response: Response): Promise<AuthState> {
  const data = (await response.json()) as AuthState;
  if (!data.policy) {
    data.policy = emptyPolicy;
  }
  return data;
}

async function readPeers(response: Response): Promise<PeersState> {
  const data = (await response.json()) as PeersState;
  data.peers ??= [];
  return data;
}

function peerText(peers?: string[]) {
  return (peers ?? []).join("\n");
}

const peerTypeLabels: Record<string, string> = {
  all: "All",
  user: "Users",
  group: "Groups",
  channel: "Channels",
  bot: "Bots",
};

function peersFromText(value: string) {
  return value
    .split(/[\s,]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function Toggle({
  label,
  value,
  onChange,
}: {
  label: string;
  value: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <Switch className="switch-row" isSelected={value} onChange={onChange} size="sm">
      <Switch.Content>{label}</Switch.Content>
    </Switch>
  );
}

function ApiSettingsPanel({
  api,
  onUpdated,
}: {
  api?: AuthAPI;
  onUpdated: (state: AuthState) => void;
}) {
  const [open, setOpen] = useState(false);
  const [appId, setAppId] = useState(api?.appId ? String(api.appId) : "");
  const [appHash, setAppHash] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setAppId(api?.appId ? String(api.appId) : "");
  }, [api?.appId]);

  const save = async (payload: {appId?: string; appHash?: string; useDefault?: boolean}) => {
    setSaving(true);
    setError("");
    try {
      const response = await fetch(apiURL("/auth/api"), {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      const next = await readState(response);
      if (!response.ok) {
        setError(next.error || "Не удалось обновить API.");
        return;
      }
      setAppHash("");
      setOpen(false);
      onUpdated(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось обновить API.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="api-panel">
      <div className="api-panel-head">
        <Button isDisabled={!api?.canEdit || saving} size="sm" type="button" variant="secondary" onClick={() => setOpen((value) => !value)}>
          Сторонний API
        </Button>
        <Chip color={api?.default ? "default" : "accent"} size="sm" variant="soft">
          {api?.default ? "Дефолтный API" : "Свой API"}
        </Chip>
      </div>

      {open && (
        <form
          className="api-form"
          onSubmit={(event) => {
            event.preventDefault();
            void save({appId, appHash});
          }}
        >
          {error && (
            <Alert status="danger">
              <Alert.Content>
                <Alert.Title>API</Alert.Title>
                <Alert.Description>{error}</Alert.Description>
              </Alert.Content>
            </Alert>
          )}
          <div className="api-fields">
            <Input
              fullWidth
              required
              inputMode="numeric"
              placeholder="App ID"
              value={appId}
              onChange={(event) => setAppId(event.currentTarget.value)}
            />
            <Input
              fullWidth
              required
              placeholder="App Hash"
              value={appHash}
              onChange={(event) => setAppHash(event.currentTarget.value)}
            />
          </div>
          <div className="api-actions">
            <Button isDisabled={saving} size="sm" type="submit">
              {saving ? "Применяю" : "Применить"}
            </Button>
            {!api?.default && (
              <Button
                isDisabled={saving}
                size="sm"
                type="button"
                variant="secondary"
                onClick={() => void save({useDefault: true})}
              >
                Дефолтный
              </Button>
            )}
          </div>
        </form>
      )}
    </section>
  );
}

function PolicyPanel({
  policy,
  onSaved,
}: {
  policy: Policy;
  onSaved: (state: AuthState) => void;
}) {
  const [draft, setDraft] = useState<Policy>(policy);
  const [allowText, setAllowText] = useState(peerText(policy.allowPeers));
  const [denyText, setDenyText] = useState(peerText(policy.denyPeers));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setDraft(policy);
    setAllowText(peerText(policy.allowPeers));
    setDenyText(peerText(policy.denyPeers));
  }, [policy]);

  const setSafety = (key: keyof Safeties, value: boolean) => {
    setDraft((current) => ({
      ...current,
      safeties: {...current.safeties, [key]: value},
    }));
  };

  const setPeerType = (key: keyof PeerTypes, value: boolean) => {
    setDraft((current) => ({
      ...current,
      peerTypes: {...current.peerTypes, [key]: value},
    }));
  };

  const save = async (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      const response = await fetch(apiURL("/auth/policy"), {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          ...draft,
          allowPeers: peersFromText(allowText),
          denyPeers: peersFromText(denyText),
        }),
      });
      const next = await readState(response);
      if (!response.ok) {
        setError(next.error || "Failed to save permissions.");
        return;
      }
      onSaved(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save permissions.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <form className="grid gap-4 border-t border-default-200 pt-5" onSubmit={save}>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 className="m-0 text-base font-semibold">Local permissions</h2>
          <p className="m-0 text-sm text-default-500">Saved before login and enforced by IPC/HTTP.</p>
        </div>
        <Button isDisabled={saving} size="sm" type="submit" variant="secondary">
          {saving ? "Saving" : "Save"}
        </Button>
      </div>

      {error && (
        <Alert status="danger">
          <Alert.Content>
            <Alert.Title>Policy error</Alert.Title>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      )}

      <section className="grid gap-2">
        <div className="text-sm font-semibold text-default-700">Actions</div>
        <div className="policy-grid">
          <Toggle label="Read" value={draft.safeties.read} onChange={(value) => setSafety("read", value)} />
          <Toggle label="Write" value={draft.safeties.write} onChange={(value) => setSafety("write", value)} />
          <Toggle
            label="Destructive"
            value={draft.safeties.destructive}
            onChange={(value) => setSafety("destructive", value)}
          />
          <Toggle label="Paid" value={draft.safeties.paid} onChange={(value) => setSafety("paid", value)} />
        </div>
      </section>

      <section className="grid gap-2">
        <div className="text-sm font-semibold text-default-700">Dialogs</div>
        <div className="policy-grid">
          <Toggle label="Private chats" value={draft.peerTypes.users} onChange={(value) => setPeerType("users", value)} />
          <Toggle label="Groups" value={draft.peerTypes.groups} onChange={(value) => setPeerType("groups", value)} />
          <Toggle label="Channels" value={draft.peerTypes.channels} onChange={(value) => setPeerType("channels", value)} />
          <Toggle label="Bots" value={draft.peerTypes.bots} onChange={(value) => setPeerType("bots", value)} />
        </div>
      </section>

      <section className="peer-grid">
        <label className="grid gap-2 text-sm font-semibold text-default-700">
          Allowed peers
          <TextArea
            fullWidth
            placeholder="@username, -100123, user:42"
            rows={4}
            value={allowText}
            onChange={(event) => setAllowText(event.currentTarget.value)}
          />
        </label>
        <label className="grid gap-2 text-sm font-semibold text-default-700">
          Denied peers
          <TextArea
            fullWidth
            placeholder="@username, bot:example"
            rows={4}
            value={denyText}
            onChange={(event) => setDenyText(event.currentTarget.value)}
          />
        </label>
      </section>
    </form>
  );
}

function PeerAccessPanel({
  policy,
  onFinished,
}: {
  policy: Policy;
  onFinished: (state: AuthState) => void;
}) {
  const [peersState, setPeersState] = useState<PeersState>({
    peers: [],
    count: 0,
    loaded: false,
    loading: true,
  });
  const [selected, setSelected] = useState<Set<string>>(() => new Set(policy.allowPeers ?? []));
  const [selectionReady, setSelectionReady] = useState(false);
  const [query, setQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const loadPeers = useCallback(async () => {
    try {
      const response = await fetch(apiURL("/auth/peers"), {
        headers: {Accept: "application/json"},
      });
      const next = await readPeers(response);
      setPeersState(next);
      if (!response.ok) {
        setError(next.error || "Failed to load dialogs.");
      } else if (!next.error) {
        setError("");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load dialogs.");
    }
  }, []);

  useEffect(() => {
    void loadPeers();
  }, [loadPeers]);

  useEffect(() => {
    if (peersState.loaded || !peersState.loading) {
      return undefined;
    }
    const id = window.setInterval(() => void loadPeers(), 1500);
    return () => window.clearInterval(id);
  }, [loadPeers, peersState.loaded, peersState.loading]);

  useEffect(() => {
    if (selectionReady || !peersState.loaded) {
      return;
    }
    const saved = policy.allowPeers ?? [];
    setSelected(new Set(saved.length > 0 ? saved : peersState.peers.map((peer) => peer.peer)));
    setSelectionReady(true);
  }, [peersState.loaded, peersState.peers, policy.allowPeers, selectionReady]);

  const filteredPeers = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    return peersState.peers.filter((peer) => {
      if (typeFilter !== "all" && peer.type !== typeFilter) {
        return false;
      }
      if (!normalizedQuery) {
        return true;
      }
      return [peer.title, peer.username, peer.peer].some((value) => value?.toLowerCase().includes(normalizedQuery));
    });
  }, [peersState.peers, query, typeFilter]);

  const selectedCount = selected.size;

  const togglePeer = (peer: string, value: boolean) => {
    setSelected((current) => {
      const next = new Set(current);
      if (value) {
        next.add(peer);
      } else {
        next.delete(peer);
      }
      return next;
    });
  };

  const setVisible = (value: boolean) => {
    setSelected((current) => {
      const next = new Set(current);
      for (const peer of filteredPeers) {
        if (value) {
          next.add(peer.peer);
        } else {
          next.delete(peer.peer);
        }
      }
      return next;
    });
  };

  const onPeerKeyDown = (event: React.KeyboardEvent<HTMLDivElement>, peer: string, value: boolean) => {
    if (event.key !== "Enter" && event.key !== " ") {
      return;
    }
    event.preventDefault();
    togglePeer(peer, value);
  };

  const finish = async () => {
    setSaving(true);
    setError("");
    const allowPeers = Array.from(selected).sort();
    const nextPolicy: Policy = {
      ...policy,
      allowPeers,
      denyPeers: policy.denyPeers ?? [],
      peerTypes:
        allowPeers.length === 0
          ? {users: false, groups: false, channels: false, bots: false}
          : policy.peerTypes,
    };

    try {
      const response = await fetch(apiURL("/auth/finish"), {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({policy: nextPolicy}),
      });
      const next = await readState(response);
      if (!response.ok) {
        setError(next.error || "Failed to save filter.");
        return;
      }
      onFinished(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save filter.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="setup-panel">
      <div className="setup-heading">
        <div>
          <h2>Кому разрешить доступ</h2>
          <p>Выбранные диалоги станут allowlist для агента.</p>
        </div>
        <Chip color="success" variant="soft">
          {selectedCount}
        </Chip>
      </div>

      {error && (
        <Alert status="danger">
          <Alert.Content>
            <Alert.Title>Filter error</Alert.Title>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      )}

      <div className="peer-toolbar">
        <Input
          fullWidth
          placeholder="Поиск по названию или username"
          value={query}
          onChange={(event) => setQuery(event.currentTarget.value)}
        />
        <div className="type-filter" role="group" aria-label="Dialog type filter">
          {Object.entries(peerTypeLabels).map(([value, label]) => (
            <Button
              key={value}
              size="sm"
              type="button"
              variant={typeFilter === value ? "primary" : "secondary"}
              onClick={() => setTypeFilter(value)}
            >
              {label}
            </Button>
          ))}
        </div>
      </div>

      <div className="setup-actions">
        <div className="selection-count">
          {peersState.loading ? "Загружаю диалоги..." : `${selectedCount} из ${peersState.count} выбрано`}
        </div>
        <div className="flex flex-wrap gap-2">
          <Button size="sm" type="button" variant="secondary" onClick={() => setVisible(true)}>
            Выбрать видимые
          </Button>
          <Button size="sm" type="button" variant="secondary" onClick={() => setVisible(false)}>
            Снять видимые
          </Button>
        </div>
      </div>

      {peersState.loading && !peersState.loaded ? (
        <div className="flex min-h-56 items-center justify-center">
          <Spinner size="lg" />
        </div>
      ) : (
        <div className="peer-list">
          {filteredPeers.map((peer) => {
            const isSelected = selected.has(peer.peer);
            return (
              <div
                aria-checked={isSelected}
                className={`peer-row ${isSelected ? "is-selected" : ""}`}
                key={peer.peer}
                role="checkbox"
                tabIndex={0}
                onClick={() => togglePeer(peer.peer, !isSelected)}
                onKeyDown={(event) => onPeerKeyDown(event, peer.peer, !isSelected)}
              >
                <span className="peer-check" aria-hidden="true" />
                <div className="peer-copy">
                  <span className="peer-title">{peer.title}</span>
                  <span className="peer-meta">{peer.username ? `@${peer.username}` : peer.peer}</span>
                </div>
                <Chip size="sm" variant="soft">
                  {peerTypeLabels[peer.type] ?? peer.type}
                </Chip>
              </div>
            );
          })}
          {filteredPeers.length === 0 && <div className="empty-peers">Ничего не найдено</div>}
        </div>
      )}

      <Button className="finish-button" isDisabled={saving} type="button" onClick={finish}>
        {saving ? "Сохраняю" : "Сохранить и завершить"}
      </Button>
    </section>
  );
}

function AuthBody({
  state,
  setState,
}: {
  state: AuthState;
  setState: (state: AuthState) => void;
}) {
  const [code, setCode] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setError(state.error ?? "");
  }, [state.error, state.mode]);

  const submit = async (path: string, payload: Record<string, string>) => {
    setSubmitting(true);
    setError("");
    try {
      const response = await fetch(apiURL(path), {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      const next = await readState(response);
      setState(next);
      if (!response.ok) {
        setError(next.error || "Authentication failed.");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed.");
    } finally {
      setSubmitting(false);
    }
  };

  if (state.mode === "setup") {
    return null;
  }

  if (state.completed || state.mode === "done") {
    return (
      <div className="grid gap-4">
        <Chip color="success" variant="soft">
          Login complete
        </Chip>
        <p className="m-0 text-default-600">You can close this page and return to the terminal.</p>
      </div>
    );
  }

  if (state.mode === "qr") {
    return (
      <div className="qr-stage">
        {state.qrImage ? (
          <img alt="Telegram QR code" className="qr-image" src={state.qrImage} />
        ) : (
          <div className="flex min-h-56 items-center justify-center">
            <Spinner size="lg" />
          </div>
        )}
      </div>
    );
  }

  if (state.mode === "password") {
    return (
      <form
        className="grid gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          void submit("/auth/password", {password});
        }}
      >
        {error && (
          <Alert status="danger">
            <Alert.Content>
              <Alert.Title>2FA failed</Alert.Title>
              <Alert.Description>{error}</Alert.Description>
            </Alert.Content>
          </Alert>
        )}
        <Input
          autoFocus
          fullWidth
          required
          placeholder="Password"
          type="password"
          value={password}
          onChange={(event) => setPassword(event.currentTarget.value)}
        />
        <Button isDisabled={submitting} type="submit">
          {submitting ? "Completing" : "Complete login"}
        </Button>
      </form>
    );
  }

  return (
    <form
      className="grid gap-3"
      onSubmit={(event) => {
        event.preventDefault();
        void submit("/auth/verify", {code});
      }}
    >
      {error && (
        <Alert status="danger">
          <Alert.Content>
            <Alert.Title>Login failed</Alert.Title>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      )}
      <Input
        autoFocus
        fullWidth
        required
        inputMode="numeric"
        placeholder="Code"
        value={code}
        onChange={(event) => setCode(event.currentTarget.value)}
      />
      <Button isDisabled={submitting} type="submit">
        {submitting ? "Verifying" : "Verify code"}
      </Button>
    </form>
  );
}

function App() {
  const [state, setState] = useState<AuthState | null>(null);
  const [loadError, setLoadError] = useState("");

  const load = useCallback(async () => {
    try {
      const response = await fetch(apiURL("/auth/state"), {
        headers: {Accept: "application/json"},
      });
      const next = await readState(response);
      if (!response.ok) {
        setLoadError(next.error || "Failed to load auth state.");
        return;
      }
      setState(next);
      setLoadError("");
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : "Failed to load auth state.");
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!state || state.completed) {
      return undefined;
    }
    const delay = Math.max(1000, (state.refresh ?? 2) * 1000);
    const id = window.setInterval(() => void load(), delay);
    return () => window.clearInterval(id);
  }, [load, state]);

  const policy = useMemo(() => state?.policy ?? emptyPolicy, [state]);
  const setupMode = state?.mode === "setup";
  const qrMode = state?.mode === "qr";

  return (
    <main className="auth-shell">
      <Card className="auth-card" variant="default">
        <Card.Header className="auth-header">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Card.Title className="auth-title">{state?.title ?? "Telegram login"}</Card.Title>
            <Chip color={state?.completed ? "success" : "accent"} size="sm" variant="soft">
              {state?.completed ? "Done" : state?.mode === "qr" ? "QR" : "Auth"}
            </Chip>
          </div>
          <Card.Description className="auth-description">
            {state?.message || state?.hint || "Preparing the local authentication page."}
          </Card.Description>
        </Card.Header>

        <Card.Content className="grid gap-5">
          {loadError && (
            <Alert status="danger">
              <Alert.Content>
                <Alert.Title>Auth page error</Alert.Title>
                <Alert.Description>{loadError}</Alert.Description>
              </Alert.Content>
            </Alert>
          )}

          {state ? (
            <>
              <AuthBody state={state} setState={setState} />
              {qrMode ? (
                <ApiSettingsPanel api={state.api} onUpdated={setState} />
              ) : setupMode ? (
                <PeerAccessPanel policy={policy} onFinished={setState} />
              ) : (
                !state.completed && <PolicyPanel policy={policy} onSaved={setState} />
              )}
            </>
          ) : (
            <div className="flex min-h-72 items-center justify-center">
              <Spinner size="lg" />
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
