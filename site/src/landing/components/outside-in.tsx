import { ArrowRight, BracketsCurly, CheckCircle, Pulse } from "@phosphor-icons/react";

import { OUTSIDE_IN_COMPARISON } from "../content";
import { SectionReveal } from "./section-reveal";

type ComparisonColumnProps = {
  accent?: boolean;
  eyebrow: string;
  items: readonly string[];
  title: string;
};

function ComparisonColumn({
  accent = false,
  eyebrow,
  items,
  title,
}: ComparisonColumnProps): React.JSX.Element {
  return (
    <div className={`outside-column${accent ? " outside-column--accent" : ""}`}>
      <div className="outside-column__label">
        {accent ? <Pulse size={17} weight="duotone" /> : <BracketsCurly size={17} weight="duotone" />}
        {eyebrow}
      </div>
      <h3>{title}</h3>
      <ul>
        {items.map((item) => (
          <li key={item}>
            <CheckCircle aria-hidden="true" size={18} weight={accent ? "fill" : "duotone"} />
            <span>{item}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function OutsideIn(): React.JSX.Element {
  return (
    <section className="section outside-section" id="testing">
      <div className="section-heading">
        <SectionReveal>
          <div className="section-kicker">
            <span>01</span>
            OUTSIDE-IN TESTING
          </div>
          <h2>Your endpoint can pass while the conversation is broken.</h2>
        </SectionReveal>
        <SectionReveal delay={0.08}>
          <p>
            Most bot tests stop at the handler: the request arrived, the payload was valid, and
            Telegram accepted the response. Users experience the complete sequence — messages,
            buttons, delays, follow-ups, and state transitions.
          </p>
        </SectionReveal>
      </div>

      <SectionReveal className="outside-card">
        <ComparisonColumn {...OUTSIDE_IN_COMPARISON.backend} />

        <div className="outside-bridge" aria-hidden="true">
          <span>REAL CHAT</span>
          <div>
            <ArrowRight size={18} weight="bold" />
          </div>
        </div>

        <ComparisonColumn {...OUTSIDE_IN_COMPARISON.experience} accent />

        <div className="outside-takeaway">
          <p>
            <strong>Use API tests to verify the implementation.</strong> Use a real-user session
            to verify the experience.
          </p>
          <div aria-label="Run artifacts" className="outside-artifacts">
            {OUTSIDE_IN_COMPARISON.artifacts.map((artifact) => (
              <span key={artifact}>{artifact}</span>
            ))}
          </div>
        </div>
      </SectionReveal>
    </section>
  );
}
