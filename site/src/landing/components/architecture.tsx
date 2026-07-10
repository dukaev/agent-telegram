import { Check, Code, Lightning } from "@phosphor-icons/react";
import gsap from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { useLayoutEffect, useRef } from "react";

import { ARCHITECTURE_STEPS } from "../content";
import { useReducedMotionPreference } from "../use-reduced-motion";
import { SectionReveal } from "./section-reveal";

gsap.registerPlugin(ScrollTrigger);

const CONTRACT_ROWS = [
  ["Output", "Structured JSON"],
  ["Discovery", "Manifest + schemas"],
  ["Safety", "read / write / destructive / paid"],
  ["Debugging", "run IDs + receipts + traces"],
] as const;

export function Architecture(): React.JSX.Element {
  const root = useRef<HTMLElement>(null);
  const reduceMotion = useReducedMotionPreference();

  useLayoutEffect(() => {
    if (!root.current) return;

    const context = gsap.context(() => {
      if (reduceMotion) {
        gsap.set(".architecture-line__fill", { scaleX: 1 });
        gsap.set(".architecture-signal", { left: "100%" });
        return;
      }

      const timeline = gsap.timeline({
        scrollTrigger: {
          end: "+=55%",
          scrub: 0.6,
          start: "top 72%",
          trigger: root.current,
        },
      });

      timeline
        .fromTo(".architecture-line__fill", { scaleX: 0 }, { ease: "none", scaleX: 1 })
        .fromTo(".architecture-signal", { left: "0%" }, { ease: "none", left: "100%" }, 0)
        .fromTo(
          ".architecture-step",
          { opacity: 0.45, y: 8 },
          { duration: 0.22, opacity: 1, stagger: 0.18, y: 0 },
          0,
        );
    }, root);

    return () => context.revert();
  }, [reduceMotion]);

  return (
    <section className="section architecture-section" id="architecture" ref={root}>
      <div className="section-heading section-heading--split">
        <SectionReveal>
          <div className="section-kicker">
            <span>03</span>
            HOW IT WORKS
          </div>
          <h2>Fast commands. Local state. Real Telegram.</h2>
        </SectionReveal>
        <SectionReveal delay={0.08}>
          <p>
            The agent sees a stable machine-readable contract. Your Telegram session and daemon
            stay on the machine you control.
          </p>
        </SectionReveal>
      </div>

      <div className="architecture-card">
        <div className="architecture-path">
          <div className="architecture-line" aria-hidden="true">
            <span className="architecture-line__fill" />
            <span className="architecture-signal">
              <Lightning size={12} weight="fill" />
            </span>
          </div>
          {ARCHITECTURE_STEPS.map(({ icon: Icon, title, detail }, index) => (
            <div className="architecture-step" key={title}>
              <div className="architecture-step__node">
                <Icon size={25} weight="duotone" />
              </div>
              <span>0{index + 1}</span>
              <strong>{title}</strong>
              <small>{detail}</small>
            </div>
          ))}
        </div>

        <div className="contract-card">
          <div className="contract-card__header">
            <Code size={18} weight="duotone" />
            AGENT CONTRACT
            <span>ready</span>
          </div>
          {CONTRACT_ROWS.map(([label, value]) => (
            <div className="contract-row" key={label}>
              <Check size={14} weight="bold" />
              <span>{label}</span>
              <strong>{value}</strong>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
