import { ArrowDown, ArrowUpRight, GithubLogo, QrCode } from "@phosphor-icons/react";
import gsap from "gsap";
import { useLayoutEffect, useRef } from "react";

import { AGENT_SETUP_PROMPT, GITHUB_URL } from "../content";
import { useReducedMotionPreference } from "../use-reduced-motion";
import { AsciiDuckVideo } from "./ascii-duck-video";
import { CopyCommand } from "./copy-command";

export function Hero(): React.JSX.Element {
  const root = useRef<HTMLElement>(null);
  const reduceMotion = useReducedMotionPreference();

  useLayoutEffect(() => {
    if (!root.current || reduceMotion) return;

    const context = gsap.context(() => {
      gsap
        .timeline({ defaults: { ease: "power3.out" } })
        .fromTo(
          "[data-hero-reveal]",
          { autoAlpha: 0, y: 26 },
          { autoAlpha: 1, duration: 0.85, stagger: 0.075, y: 0 },
        )
        .fromTo(
          "[data-demo-reveal]",
          { autoAlpha: 0, scale: 0.97, y: 34 },
          { autoAlpha: 1, duration: 1.05, scale: 1, y: 0 },
          "-=0.62",
        );
    }, root);

    return () => context.revert();
  }, [reduceMotion]);

  return (
    <section className="hero" id="top" ref={root}>
      <div className="hero__copy">
        <div className="eyebrow" data-hero-reveal>
          <span className="eyebrow__pulse" />
          TELEGRAM FOR AI AGENTS
          <span className="eyebrow__line" />
          LOCAL-FIRST
        </div>

        <h1 data-hero-reveal>
          Your agent can use Telegram. <em>Like a real user.</em>
        </h1>

        <p className="hero__lede" data-hero-reveal>
          Paste one prompt into your coding agent. It handles setup and opens Telegram sign-in.
          Scan the QR code once — you&apos;re ready.
        </p>

        <div data-hero-reveal>
          <CopyCommand command={AGENT_SETUP_PROMPT} variant="prompt" />
        </div>

        <div className="hero__qr-cue" data-hero-reveal>
          <span>02</span>
          <QrCode aria-hidden="true" size={22} weight="duotone" />
          <p>
            <strong>Scan one QR code.</strong> That&apos;s the entire sign-in.
          </p>
        </div>

        <div className="hero__actions" data-hero-reveal>
          <a className="button button--primary" href="#install">
            <QrCode size={18} weight="bold" />
            See the 2-step setup
          </a>
          <a className="button button--ghost" href={GITHUB_URL} rel="noreferrer" target="_blank">
            <GithubLogo size={19} weight="fill" />
            View on GitHub
            <ArrowUpRight size={16} />
          </a>
        </div>

        <div className="hero__microproof" data-hero-reveal>
          <span>Agent installs it</span>
          <span>One QR scan</span>
          <span>Session stays local</span>
        </div>
      </div>

      <div className="hero__ascii-wrap" data-demo-reveal>
        <div className="hero__duck-stage">
          <AsciiDuckVideo />
        </div>
      </div>

      <a aria-label="Scroll to product proof" className="scroll-cue" href="#proof">
        <ArrowDown size={16} />
        Scroll to inspect
      </a>
    </section>
  );
}
