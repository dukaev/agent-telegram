import { GithubLogo } from "@phosphor-icons/react";
import { useEffect, useRef } from "react";

import { Architecture } from "../landing/components/architecture";
import { FinalCta } from "../landing/components/final-cta";
import { Hero } from "../landing/components/hero";
import { OutsideIn } from "../landing/components/outside-in";
import { ProofStrip } from "../landing/components/proof-strip";
import { Safety } from "../landing/components/safety";
import { SiteHeader } from "../landing/components/site-header";
import { UseCases } from "../landing/components/use-cases";
import { GITHUB_URL } from "../landing/content";

function PointerSpotlight(): React.JSX.Element {
  const glow = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const element = glow.current;
    const finePointer = window.matchMedia("(hover: hover) and (pointer: fine)").matches;
    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (!element || !finePointer || reduced) return;

    let frame: number | null = null;
    const move = (event: PointerEvent) => {
      if (frame !== null) cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => {
        element.style.transform = `translate3d(${event.clientX - 240}px, ${event.clientY - 240}px, 0)`;
        element.dataset.visible = "true";
      });
    };
    const hide = () => {
      element.dataset.visible = "false";
    };

    window.addEventListener("pointermove", move, { passive: true });
    document.documentElement.addEventListener("mouseleave", hide);
    return () => {
      if (frame !== null) cancelAnimationFrame(frame);
      window.removeEventListener("pointermove", move);
      document.documentElement.removeEventListener("mouseleave", hide);
    };
  }, []);

  return <div aria-hidden="true" className="pointer-spotlight" ref={glow} />;
}

export function AppHome(): React.JSX.Element {
  return (
    <div className="site-shell">
      <PointerSpotlight />
      <div className="ambient-grid" aria-hidden="true" />
      <SiteHeader />
      <main>
        <Hero />
        <ProofStrip />
        <OutsideIn />
        <UseCases />
        <Architecture />
        <Safety />
        <FinalCta />
      </main>
      <footer className="site-footer">
        <a className="brand" href="#top">
          <span className="brand__mark" aria-hidden="true">
            <span />
          </span>
          <span>agent-telegram</span>
        </a>
        <p>Telegram automation CLI for AI agents.</p>
        <div>
          <span>MIT · LOCAL-FIRST</span>
          <a href={GITHUB_URL} rel="noreferrer" target="_blank">
            <GithubLogo size={17} weight="fill" /> GitHub
          </a>
        </div>
      </footer>
    </div>
  );
}
