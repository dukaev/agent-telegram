import { ArrowsClockwise, Check, Pause, Play } from "@phosphor-icons/react";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useMemo, useState } from "react";

import type { BotScenario } from "../content";
import { useReducedMotionPreference } from "../use-reduced-motion";

type TelegramDemoProps = {
  scenario: BotScenario;
  compact?: boolean;
};

export function TelegramDemo({ scenario, compact = false }: TelegramDemoProps): React.JSX.Element {
  const reduceMotion = useReducedMotionPreference();
  const [visibleEvents, setVisibleEvents] = useState(reduceMotion ? scenario.events.length : 1);
  const [playing, setPlaying] = useState(!reduceMotion);
  const progress = visibleEvents / scenario.events.length;

  useEffect(() => {
    setVisibleEvents(reduceMotion ? scenario.events.length : 1);
    setPlaying(!reduceMotion);
  }, [reduceMotion, scenario.id, scenario.events.length]);

  useEffect(() => {
    if (!playing || reduceMotion || visibleEvents >= scenario.events.length) return;

    const timer = window.setTimeout(
      () => setVisibleEvents((current) => Math.min(current + 1, scenario.events.length)),
      compact ? 820 : 1_080,
    );
    return () => window.clearTimeout(timer);
  }, [compact, playing, reduceMotion, scenario.events.length, visibleEvents]);

  const status = useMemo(() => {
    if (visibleEvents >= scenario.events.length) return "passed";
    return playing ? "running" : "paused";
  }, [playing, scenario.events.length, visibleEvents]);

  const restart = () => {
    setVisibleEvents(reduceMotion ? scenario.events.length : 1);
    setPlaying(!reduceMotion);
  };

  return (
    <div
      aria-label={`${scenario.title} interactive demo`}
      className={`telegram-demo${compact ? " telegram-demo--compact" : ""}`}
      data-autoplay={playing && !reduceMotion ? "true" : "false"}
      data-demo-scenario={scenario.id}
    >
      <div className="telegram-demo__topbar">
        <div className="window-dots" aria-hidden="true">
          <span />
          <span />
          <span />
        </div>
        <div className={`run-status run-status--${status}`}>
          <span aria-hidden="true" />
          {status === "passed" ? "run passed" : status}
        </div>
        <div className="demo-controls">
          <button aria-label="Restart demo" onClick={restart} type="button">
            <ArrowsClockwise size={15} />
          </button>
          <button
            aria-label={playing ? "Pause demo" : "Play demo"}
            disabled={reduceMotion || visibleEvents >= scenario.events.length}
            onClick={() => setPlaying((current) => !current)}
            type="button"
          >
            {playing ? <Pause size={14} weight="fill" /> : <Play size={14} weight="fill" />}
          </button>
        </div>
      </div>

      <div className="telegram-demo__body">
        <div className="terminal-pane">
          <div className="pane-label">
            <span>AGENT RUN</span>
            <span>run_ci_284</span>
          </div>
          <div className="terminal-command">
            <span className="prompt">$</span>
            <code>{scenario.command}</code>
          </div>
          <div className="terminal-log" aria-live="polite">
            {scenario.events.slice(0, visibleEvents).map((event, index) => (
              <motion.div
                animate={{ opacity: 1, transform: "translateY(0)" }}
                className="terminal-log__row"
                data-demo-step={index + 1}
                initial={reduceMotion ? false : { opacity: 0, transform: "translateY(7px)" }}
                key={`${scenario.id}-${index}`}
                transition={{ duration: 0.22, ease: [0.23, 1, 0.32, 1] }}
              >
                <span>{String(index + 1).padStart(2, "0")}</span>
                <span className={`log-dot log-dot--${event.actor}`} />
                <span>{event.actor === "bot" ? "response received" : event.text.toLowerCase()}</span>
              </motion.div>
            ))}
          </div>
          <AnimatePresence>
            {status === "passed" ? (
              <motion.div
                animate={{ opacity: 1, transform: "translateY(0) scale(1)" }}
                className="terminal-result"
                initial={reduceMotion ? false : { opacity: 0, transform: "translateY(6px) scale(.98)" }}
                transition={{ type: "spring", duration: 0.45, bounce: 0.15 }}
              >
                <Check size={14} weight="bold" />
                {scenario.assertion}
              </motion.div>
            ) : null}
          </AnimatePresence>
        </div>

        <div className="chat-pane">
          <div className="chat-header">
            <div className="bot-avatar" aria-hidden="true">
              AT
            </div>
            <div>
              <strong>Acme Test Bot</strong>
              <span>bot · online</span>
            </div>
          </div>
          <div className="chat-messages" aria-live="polite">
            <AnimatePresence initial={false}>
              {scenario.events.slice(0, visibleEvents).map((event, index) =>
                event.actor === "system" ? (
                  <motion.div
                    animate={{ opacity: 1 }}
                    className="system-message"
                    initial={reduceMotion ? false : { opacity: 0 }}
                    key={`${scenario.id}-chat-${index}`}
                  >
                    {event.text} · {event.meta}
                  </motion.div>
                ) : (
                  <motion.div
                    animate={{ opacity: 1, transform: "translateY(0) scale(1)" }}
                    className={`chat-bubble chat-bubble--${event.actor}`}
                    initial={
                      reduceMotion ? false : { opacity: 0, transform: "translateY(9px) scale(.97)" }
                    }
                    key={`${scenario.id}-chat-${index}`}
                    transition={{ type: "spring", duration: 0.44, bounce: 0.12 }}
                  >
                    <span>{event.text}</span>
                    <small>{event.meta}</small>
                  </motion.div>
                ),
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      <div className="telegram-demo__progress" aria-hidden="true">
        <motion.span
          animate={{ transform: `scaleX(${progress})` }}
          initial={false}
          transition={{ duration: reduceMotion ? 0 : 0.35, ease: [0.23, 1, 0.32, 1] }}
        />
      </div>
    </div>
  );
}
