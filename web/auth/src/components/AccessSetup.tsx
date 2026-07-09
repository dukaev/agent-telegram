import React, {useCallback, useEffect, useMemo, useState} from "react";
import {Alert, Button, Chip, Input, Spinner} from "@heroui/react";

import {fetchPeers, postAuthState} from "../api";
import {AuthState, PeerOption, PeersState, peerTypeLabels, Policy} from "../types";

type AccessPreset = "all" | "bots" | "selected";

function peerInitials(peer: PeerOption) {
  return peer.title
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
}

export function AccessSetup({state, onState}: {state: AuthState; onState: (state: AuthState) => void}) {
  const policy = state.policy;
  const [peersState, setPeersState] = useState<PeersState>({peers: [], count: 0, loaded: false, loading: true});
  const [preset, setPreset] = useState<AccessPreset | null>(null);
  const [selected, setSelected] = useState<Set<string>>(() => new Set(policy.allowPeers ?? []));
  const [query, setQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const loadPeers = useCallback(async () => {
    try {
      const result = await fetchPeers();
      setPeersState(result.state);
      if (!result.ok) {
        setError(result.state.error || "Не удалось загрузить диалоги.");
      } else if (!result.state.error) {
        setError("");
      }
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Не удалось загрузить диалоги.");
    }
  }, []);

  useEffect(() => {
    void loadPeers();
  }, [loadPeers]);

  useEffect(() => {
    if (peersState.loaded || !peersState.loading) {
      return undefined;
    }
    const timer = window.setTimeout(() => void loadPeers(), 1200);
    return () => window.clearTimeout(timer);
  }, [loadPeers, peersState]);

  useEffect(() => {
    if (preset === "bots" && peersState.loaded) {
      setSelected(new Set(peersState.peers.filter((peer) => peer.type === "bot").map((peer) => peer.peer)));
    }
  }, [peersState.loaded, peersState.peers, preset]);

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

  const choosePreset = (next: AccessPreset) => {
    setPreset(next);
    setError("");
    if (next === "all") {
      setSelected(new Set());
      return;
    }
    if (next === "bots") {
      setSelected(new Set(peersState.peers.filter((peer) => peer.type === "bot").map((peer) => peer.peer)));
      return;
    }
    setSelected(new Set(policy.allowPeers ?? []));
  };

  const togglePeer = (peer: string, value: boolean) => {
    setPreset("selected");
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
    setPreset("selected");
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

  const finish = async () => {
    if (!preset) {
      setError("Сначала выбери вариант доступа.");
      return;
    }
    if (preset !== "all" && selected.size === 0) {
      setError("Выбери хотя бы один диалог или разреши доступ ко всем.");
      return;
    }

    setSaving(true);
    setError("");
    const selectedPeers = Array.from(selected).sort();
    const peerTypes = preset === "bots"
      ? {users: false, groups: false, channels: false, bots: true}
      : {users: true, groups: true, channels: true, bots: true};
    const nextPolicy: Policy = {
      ...policy,
      allowPeers: preset === "all" ? [] : selectedPeers,
      denyPeers: policy.denyPeers ?? [],
      peerTypes,
    };

    try {
      const result = await postAuthState("/auth/finish", {policy: nextPolicy});
      onState(result.state);
      if (!result.ok) {
        setError(result.state.error || "Не удалось сохранить доступ.");
      }
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Не удалось сохранить доступ.");
    } finally {
      setSaving(false);
    }
  };

  const selectionLabel = preset === "all"
    ? "Все текущие и будущие диалоги"
    : preset
      ? `${selected.size} из ${peersState.count} диалогов`
      : "Доступ ещё не выбран";

  return (
    <section className="access-setup">
      {error && (
        <Alert status="danger" role="alert">
          <Alert.Content>
            <Alert.Title>Проверь настройку доступа</Alert.Title>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      )}

      <div className="preset-grid" role="radiogroup" aria-label="Вариант доступа к диалогам">
        <button aria-checked={preset === "all"} className={`preset-card ${preset === "all" ? "is-selected" : ""}`} role="radio" type="button" onClick={() => choosePreset("all")}>
          <span className="preset-icon" aria-hidden="true">∞</span>
          <span><strong>Все диалоги</strong><small>Текущие и новые чаты</small></span>
          <span className="preset-radio" aria-hidden="true" />
        </button>
        <button aria-checked={preset === "bots"} className={`preset-card ${preset === "bots" ? "is-selected" : ""}`} role="radio" type="button" onClick={() => choosePreset("bots")}>
          <span className="preset-icon" aria-hidden="true">BOT</span>
          <span><strong>Только боты</strong><small>Без личных чатов, групп и каналов</small></span>
          <span className="preset-radio" aria-hidden="true" />
        </button>
        <button aria-checked={preset === "selected"} className={`preset-card ${preset === "selected" ? "is-selected" : ""}`} role="radio" type="button" onClick={() => choosePreset("selected")}>
          <span className="preset-icon" aria-hidden="true">✓</span>
          <span><strong>Выбрать вручную</strong><small>Точный allowlist диалогов</small></span>
          <span className="preset-radio" aria-hidden="true" />
        </button>
      </div>

      {preset === "all" ? (
        <div className="access-summary">
          <strong>Агент сможет работать со всеми диалогами</strong>
          <p>Операции удаления и оплаты по-прежнему требуют отдельного подтверждения.</p>
        </div>
      ) : preset && (
        <div className="peer-picker">
          <div className="peer-toolbar">
            <Input
              aria-label="Поиск диалогов"
              fullWidth
              placeholder="Поиск по названию или username"
              value={query}
              onChange={(event) => setQuery(event.currentTarget.value)}
            />
            <div className="type-filter" role="group" aria-label="Тип диалога">
              {Object.entries(peerTypeLabels).map(([value, label]) => (
                <Button key={value} size="sm" type="button" variant={typeFilter === value ? "primary" : "secondary"} onClick={() => setTypeFilter(value)}>
                  {label}
                </Button>
              ))}
            </div>
          </div>

          <div className="selection-toolbar">
            <span>{peersState.loading ? "Загружаю диалоги…" : `${selected.size} из ${peersState.count} выбрано`}</span>
            <div>
              <button type="button" onClick={() => setVisible(true)}>Выбрать видимые</button>
              <button type="button" onClick={() => setVisible(false)}>Снять видимые</button>
            </div>
          </div>

          {peersState.loading && !peersState.loaded ? (
            <div className="peer-loading"><Spinner size="lg" /><span>Получаю список диалогов…</span></div>
          ) : (
            <div className="peer-list">
              {filteredPeers.map((peer) => {
                const isSelected = selected.has(peer.peer);
                return (
                  <div
                    aria-checked={isSelected}
                    aria-label={`${peer.title}, ${peerTypeLabels[peer.type] ?? peer.type}`}
                    className={`peer-row ${isSelected ? "is-selected" : ""}`}
                    key={peer.peer}
                    role="checkbox"
                    tabIndex={0}
                    onClick={() => togglePeer(peer.peer, !isSelected)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault();
                        togglePeer(peer.peer, !isSelected);
                      }
                    }}
                  >
                    <span className={`peer-avatar peer-avatar--${peer.type}`} aria-hidden="true">{peerInitials(peer)}</span>
                    <span className="peer-copy">
                      <strong>{peer.title}</strong>
                      <small>{peer.username ? `@${peer.username}` : peer.peer}</small>
                    </span>
                    <Chip size="sm" variant="soft">{peerTypeLabels[peer.type] ?? peer.type}</Chip>
                    <span className="peer-check" aria-hidden="true" />
                  </div>
                );
              })}
              {filteredPeers.length === 0 && <div className="empty-peers">По этому запросу ничего не найдено</div>}
            </div>
          )}
        </div>
      )}

      <div className="finish-bar">
        <div><small>Будет сохранено</small><strong>{selectionLabel}</strong></div>
        <Button isDisabled={saving || !preset || (preset !== "all" && selected.size === 0)} type="button" onClick={() => void finish()}>
          {saving ? "Сохраняю…" : "Сохранить и завершить"}
        </Button>
      </div>
    </section>
  );
}
