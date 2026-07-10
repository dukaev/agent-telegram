import { PROOF_POINTS } from "../content";
import { SectionReveal } from "./section-reveal";

export function ProofStrip(): React.JSX.Element {
  return (
    <section aria-label="Product proof" className="proof-strip" id="proof">
      <SectionReveal className="proof-strip__inner">
        {PROOF_POINTS.map(({ icon: Icon, label, value }) => (
          <div className="proof-item" key={label}>
            <Icon aria-hidden="true" size={21} weight="duotone" />
            <div>
              <span>{label}</span>
              <strong>{value}</strong>
            </div>
          </div>
        ))}
      </SectionReveal>
    </section>
  );
}
