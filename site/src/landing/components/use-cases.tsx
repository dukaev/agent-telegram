import { ArrowUpRight } from "@phosphor-icons/react";
import { useRef } from "react";

import { USE_CASES, type UseCase } from "../content";
import { SectionReveal } from "./section-reveal";

function UseCaseCard({ item, index }: { item: UseCase; index: number }): React.JSX.Element {
  const cardRef = useRef<HTMLElement>(null);
  const frame = useRef<number | null>(null);

  const reset = () => {
    if (frame.current !== null) cancelAnimationFrame(frame.current);
    if (cardRef.current) cardRef.current.style.transform = "perspective(900px) rotateX(0) rotateY(0)";
  };

  return (
    <SectionReveal delay={index * 0.055}>
      <article
        className={`use-case-card use-case-card--${item.accent}`}
        onPointerLeave={reset}
        onPointerMove={(event) => {
          if (event.pointerType !== "mouse" || !cardRef.current) return;
          const card = cardRef.current;
          const rect = card.getBoundingClientRect();
          const x = (event.clientX - rect.left) / rect.width - 0.5;
          const y = (event.clientY - rect.top) / rect.height - 0.5;

          if (frame.current !== null) cancelAnimationFrame(frame.current);
          frame.current = requestAnimationFrame(() => {
            card.style.transform = `perspective(900px) rotateX(${-y * 4}deg) rotateY(${x * 5}deg)`;
            card.style.setProperty("--pointer-x", `${(x + 0.5) * 100}%`);
            card.style.setProperty("--pointer-y", `${(y + 0.5) * 100}%`);
          });
        }}
        ref={cardRef}
        tabIndex={0}
      >
        <div className="use-case-card__shine" aria-hidden="true" />
        <div className="use-case-card__header">
          <span>{item.number}</span>
          <item.icon size={28} weight="duotone" />
        </div>
        <div>
          <h3>{item.title}</h3>
          <p>{item.description}</p>
        </div>
        <div className="use-case-card__footer">
          <div className="tag-list">
            {item.tags.map((tag) => (
              <span key={tag}>{tag}</span>
            ))}
          </div>
          <ArrowUpRight size={19} />
        </div>
      </article>
    </SectionReveal>
  );
}

export function UseCases(): React.JSX.Element {
  return (
    <section className="section" id="use-cases">
      <div className="section-heading section-heading--split">
        <SectionReveal>
          <div className="section-kicker">
            <span>02</span>
            BEYOND TESTING
          </div>
          <h2>One Telegram session. Many agent workflows.</h2>
        </SectionReveal>
        <SectionReveal delay={0.08}>
          <p>
            A composable tool for any workflow that needs real Telegram context — with explicit
            safety boundaries around actions that matter.
          </p>
        </SectionReveal>
      </div>

      <div className="use-case-grid">
        {USE_CASES.map((item, index) => (
          <UseCaseCard index={index} item={item} key={item.number} />
        ))}
      </div>
    </section>
  );
}
