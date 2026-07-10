import { ArrowUpRight, GithubLogo, PaperPlaneTilt, Robot, TerminalWindow } from "@phosphor-icons/react";

import {
  AGENT_SETUP_PROMPT,
  GITHUB_URL,
  INSTALL_COMMAND,
  SKILL_INSTALL_COMMAND,
} from "../content";
import { CopyCommand } from "./copy-command";
import { SectionReveal } from "./section-reveal";

export function FinalCta(): React.JSX.Element {
  return (
    <section className="section final-section" id="install">
      <SectionReveal className="final-card">
        <div className="final-card__beam" aria-hidden="true" />
        <div className="final-card__icon" aria-hidden="true">
          <PaperPlaneTilt size={36} weight="fill" />
        </div>
        <div className="section-kicker section-kicker--centered">TWO COMMANDS · NO CONFIG MAZE</div>
        <h2>Install the CLI. Add the skill. You&apos;re in.</h2>
        <p>Give your agent one prompt to install the CLI, add the skill, and verify the setup.</p>
        <div className="final-card__steps" aria-label="Two-step setup">
          <div>
            <span>01</span>
            <TerminalWindow aria-hidden="true" size={18} weight="duotone" />
            Install agent-telegram
          </div>
          <div>
            <span>02</span>
            <Robot aria-hidden="true" size={18} weight="duotone" />
            Add the agent skill
          </div>
        </div>
        <CopyCommand command={AGENT_SETUP_PROMPT} variant="prompt" />
        <p className="final-card__manual">
          Prefer manual setup? <code>{INSTALL_COMMAND}</code> · <code>{SKILL_INSTALL_COMMAND}</code>
        </p>
        <div className="final-card__links">
          <a className="button button--primary" href={GITHUB_URL} rel="noreferrer" target="_blank">
            <GithubLogo size={19} weight="fill" />
            Open GitHub
            <ArrowUpRight size={16} />
          </a>
          <a className="button button--ghost" href={`${GITHUB_URL}#quick-start`} target="_blank">
            Read quick start
          </a>
        </div>
      </SectionReveal>
    </section>
  );
}
