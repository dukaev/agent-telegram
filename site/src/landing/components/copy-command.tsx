import { Check, Copy, Robot, TerminalWindow } from "@phosphor-icons/react";
import { AnimatePresence, motion } from "motion/react";

import { useCopyCommand } from "../use-copy-command";

type CopyCommandProps = {
  command: string;
  compact?: boolean;
  variant?: "command" | "prompt";
};

export function CopyCommand({
  command,
  compact = false,
  variant = "command",
}: CopyCommandProps): React.JSX.Element {
  const { copied, copy } = useCopyCommand(command);
  const isPrompt = variant === "prompt";

  return (
    <button
      aria-label={
        copied
          ? isPrompt
            ? "Agent setup prompt copied"
            : "Install command copied"
          : isPrompt
            ? "Copy agent setup prompt"
            : `Copy install command: ${command}`
      }
      className={`copy-command${compact ? " copy-command--compact" : ""}${isPrompt ? " copy-command--prompt" : ""}`}
      onClick={() => void copy()}
      type="button"
    >
      {isPrompt ? (
        <Robot aria-hidden="true" className="copy-command__lead-icon" size={20} weight="duotone" />
      ) : (
        <TerminalWindow aria-hidden="true" size={compact ? 15 : 18} weight="duotone" />
      )}
      <span className="copy-command__body">
        {isPrompt ? <span className="copy-command__label">01 / PASTE INTO YOUR AGENT</span> : null}
        <code>{command}</code>
      </span>
      <span className="copy-command__action" aria-hidden="true">
        <AnimatePresence initial={false} mode="wait">
          <motion.span
            animate={{ filter: "blur(0px)", opacity: 1, transform: "scale(1)" }}
            className="copy-command__icon"
            exit={{ filter: "blur(2px)", opacity: 0, transform: "scale(.94)" }}
            initial={{ filter: "blur(2px)", opacity: 0, transform: "scale(.94)" }}
            key={copied ? "copied" : "copy"}
            transition={{ duration: 0.16, ease: [0.23, 1, 0.32, 1] }}
          >
            {copied ? <Check size={16} weight="bold" /> : <Copy size={16} />}
          </motion.span>
        </AnimatePresence>
      </span>
    </button>
  );
}
