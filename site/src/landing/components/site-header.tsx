import { ArrowUpRight, GithubLogo, List, X } from "@phosphor-icons/react";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useState } from "react";

import { GITHUB_URL } from "../content";

const NAV_ITEMS = [
  ["Bot testing", "#testing"],
  ["Use cases", "#use-cases"],
  ["How it works", "#architecture"],
  ["Safety", "#safety"],
] as const;

export function SiteHeader(): React.JSX.Element {
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    const closeOnResize = () => {
      if (window.innerWidth > 760) setMenuOpen(false);
    };
    window.addEventListener("resize", closeOnResize);
    return () => window.removeEventListener("resize", closeOnResize);
  }, []);

  return (
    <header className="site-header" data-hero-item>
      <a aria-label="agent-telegram home" className="brand" href="#top">
        <span className="brand__mark" aria-hidden="true">
          <span />
        </span>
        <span>agent-telegram</span>
        <span className="brand__version">v0.1</span>
      </a>

      <nav aria-label="Main navigation" className="desktop-nav">
        {NAV_ITEMS.map(([label, href]) => (
          <a href={href} key={href}>
            {label}
          </a>
        ))}
      </nav>

      <div className="site-header__actions">
        <a className="github-link" href={GITHUB_URL} rel="noreferrer" target="_blank">
          <GithubLogo size={18} weight="fill" />
          <span>GitHub</span>
        </a>
        <a className="button button--small" href="#install">
          Install
          <ArrowUpRight size={15} weight="bold" />
        </a>
        <button
          aria-expanded={menuOpen}
          aria-label={menuOpen ? "Close navigation" : "Open navigation"}
          className="menu-toggle"
          onClick={() => setMenuOpen((open) => !open)}
          type="button"
        >
          {menuOpen ? <X size={20} /> : <List size={20} />}
        </button>
      </div>

      <AnimatePresence>
        {menuOpen ? (
          <motion.nav
            animate={{ opacity: 1, transform: "translateY(0) scale(1)" }}
            aria-label="Mobile navigation"
            className="mobile-nav"
            exit={{ opacity: 0, transform: "translateY(-6px) scale(.98)" }}
            initial={{ opacity: 0, transform: "translateY(-6px) scale(.98)" }}
            transition={{ duration: 0.18, ease: [0.23, 1, 0.32, 1] }}
          >
            {NAV_ITEMS.map(([label, href]) => (
              <a href={href} key={href} onClick={() => setMenuOpen(false)}>
                {label}
                <ArrowUpRight size={16} />
              </a>
            ))}
          </motion.nav>
        ) : null}
      </AnimatePresence>
    </header>
  );
}
