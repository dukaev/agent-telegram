import { ArrowRight, BracketsCurly } from "@phosphor-icons/react";
import { AnimatePresence, motion } from "motion/react";
import { useState } from "react";

import { BOT_SCENARIOS } from "../content";
import { SectionReveal } from "./section-reveal";
import { TelegramDemo } from "./telegram-demo";

export function ScenarioLab(): React.JSX.Element {
  const [activeId, setActiveId] = useState(BOT_SCENARIOS[0].id);
  const activeScenario =
    BOT_SCENARIOS.find((scenario) => scenario.id === activeId) ?? BOT_SCENARIOS[0];

  return (
    <section className="section scenario-section" id="testing">
      <div className="section-heading">
        <SectionReveal>
          <div className="section-kicker">
            <span>01</span>
            BOT TESTING
          </div>
          <h2>Test the experience, not just the endpoint.</h2>
        </SectionReveal>
        <SectionReveal delay={0.08}>
          <p>
            Most tests stop before Telegram. agent-telegram walks the real conversation — from
            the first command to the last button and asynchronous reply.
          </p>
        </SectionReveal>
      </div>

      <div className="scenario-lab">
        <div className="scenario-tabs" role="tablist" aria-label="Bot testing scenarios">
          {BOT_SCENARIOS.map((scenario) => {
            const active = scenario.id === activeId;
            return (
              <button
                aria-controls="scenario-panel"
                aria-selected={active}
                className="scenario-tab"
                id={`scenario-tab-${scenario.id}`}
                key={scenario.id}
                onClick={() => setActiveId(scenario.id)}
                role="tab"
                type="button"
              >
                {active ? (
                  <motion.span
                    className="scenario-tab__active"
                    layoutId="scenario-active"
                    transition={{ type: "spring", duration: 0.45, bounce: 0.14 }}
                  />
                ) : null}
                <span className="scenario-tab__eyebrow">{scenario.eyebrow}</span>
                <strong>{scenario.title}</strong>
                <ArrowRight aria-hidden="true" size={17} />
              </button>
            );
          })}
        </div>

        <div
          aria-labelledby={`scenario-tab-${activeScenario.id}`}
          className="scenario-panel"
          id="scenario-panel"
          role="tabpanel"
        >
          <AnimatePresence mode="wait">
            <motion.div
              animate={{ filter: "blur(0px)", opacity: 1, transform: "translateY(0)" }}
              className="scenario-panel__copy"
              exit={{ filter: "blur(2px)", opacity: 0, transform: "translateY(-6px)" }}
              initial={{ filter: "blur(2px)", opacity: 0, transform: "translateY(8px)" }}
              key={activeScenario.id}
              transition={{ duration: 0.2, ease: [0.23, 1, 0.32, 1] }}
            >
              <div className="scenario-panel__icon">
                <BracketsCurly size={22} weight="duotone" />
              </div>
              <p>{activeScenario.description}</p>
              <code>{activeScenario.command}</code>
            </motion.div>
          </AnimatePresence>
          <TelegramDemo compact scenario={activeScenario} />
        </div>
      </div>
    </section>
  );
}
