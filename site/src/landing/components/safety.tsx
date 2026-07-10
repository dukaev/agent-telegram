import { ShieldCheck } from "@phosphor-icons/react";

import { SAFETY_POINTS } from "../content";
import { SectionReveal } from "./section-reveal";

export function Safety(): React.JSX.Element {
  return (
    <section className="section safety-section" id="safety">
      <SectionReveal className="safety-card">
        <div className="safety-card__copy">
          <div className="section-kicker section-kicker--light">
            <span>04</span>
            SAFETY IS PART OF THE API
          </div>
          <h2>Powerful enough to act. Designed to stop first.</h2>
          <p>
            agent-telegram makes risk visible to the agent. Sensitive operations are typed,
            auditable, and blocked until the right confirmation arrives.
          </p>
          <div className="safety-seal">
            <ShieldCheck size={24} weight="duotone" />
            <div>
              <strong>Local-first by default</strong>
              <span>HTTP binds to loopback unless you expose it deliberately.</span>
            </div>
          </div>
        </div>

        <div className="safety-grid">
          {SAFETY_POINTS.map(({ icon: Icon, label, value }, index) => (
            <div className="safety-point" key={label}>
              <div>
                <Icon size={22} weight="duotone" />
                <span>0{index + 1}</span>
              </div>
              <small>{label}</small>
              <strong>{value}</strong>
            </div>
          ))}
        </div>
      </SectionReveal>
    </section>
  );
}
