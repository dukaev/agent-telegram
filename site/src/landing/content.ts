import type { Icon } from "@phosphor-icons/react";
import {
  BracketsCurly,
  ChatsCircle,
  CheckCircle,
  CursorClick,
  EyeSlash,
  Key,
  LockKey,
  PlugsConnected,
  Pulse,
  Robot,
  ShieldCheck,
  TerminalWindow,
} from "@phosphor-icons/react";

export const INSTALL_COMMAND = "npm install -g agent-telegram";
export const SKILL_INSTALL_COMMAND = "agent-telegram skills install agent-telegram";
export const GITHUB_URL = "https://github.com/dukaev/agent-telegram";
export const AGENT_SETUP_PROMPT =
  "Install agent-telegram on this machine with npm install -g agent-telegram. Then run agent-telegram skills install agent-telegram. Verify the installation with a read-only check. Ask before sending messages or running paid or destructive actions.";

export type DemoEvent = {
  actor: "agent" | "bot" | "system";
  text: string;
  meta?: string;
};

export type BotScenario = {
  id: string;
  eyebrow: string;
  title: string;
  description: string;
  command: string;
  events: DemoEvent[];
  assertion: string;
};

export const BOT_SCENARIOS: BotScenario[] = [
  {
    id: "onboarding",
    eyebrow: "01 / ONBOARDING",
    title: "Run the whole /start flow",
    description:
      "Send commands, read the real reply, inspect buttons, and move through the same path as a new user.",
    command: 'agent-telegram bot step @acme_bot --send "/start" --agent',
    events: [
      { actor: "system", text: "Session ready", meta: "keychain · local" },
      { actor: "agent", text: "/start", meta: "sent by agent" },
      { actor: "bot", text: "Welcome to Acme. What do you want to build?", meta: "@acme_bot" },
      { actor: "agent", text: "Pressed “Create project”", meta: "button[0]" },
      { actor: "bot", text: "Project created. Send me a name.", meta: "reply in 428ms" },
    ],
    assertion: "Expected next action: send_text · PASS",
  },
  {
    id: "buttons",
    eyebrow: "02 / KEYBOARDS",
    title: "Press real Telegram buttons",
    description:
      "Exercise inline and reply keyboards by label or index, then verify the bot’s next state.",
    command: "agent-telegram bot press @acme_bot 4821 0 --agent",
    events: [
      { actor: "system", text: "Keyboard discovered", meta: "3 actions" },
      { actor: "bot", text: "Choose a plan", meta: "Starter · Pro · Team" },
      { actor: "agent", text: "Pressed “Pro”", meta: "inline button[1]" },
      { actor: "bot", text: "Pro selected. Continue to checkout?", meta: "@acme_bot" },
      { actor: "system", text: "Paid action paused", meta: "confirmation required" },
    ],
    assertion: "Safety gate: paid · CONFIRMATION REQUIRED",
  },
  {
    id: "regression",
    eyebrow: "03 / REGRESSION",
    title: "Wait, assert, and trace failures",
    description:
      "Keep a run ID across the scenario, wait for asynchronous replies, and inspect exactly where a regression happened.",
    command: "agent-telegram msg wait @acme_bot --after-id 4821 --timeout 20s --agent",
    events: [
      { actor: "system", text: "Run linked", meta: "run_ci_284" },
      { actor: "agent", text: "Waiting for reply…", meta: "after message 4821" },
      { actor: "bot", text: "Deployment finished successfully.", meta: "reply in 1.2s" },
      { actor: "system", text: "Receipt captured", meta: "trace_8f21" },
      { actor: "system", text: "Assertion matched", meta: "status = success" },
    ],
    assertion: "Trace complete · 5 operations · PASS",
  },
];

export type FeaturePoint = {
  icon: Icon;
  label: string;
  value: string;
};

export const PROOF_POINTS: FeaturePoint[] = [
  { icon: Robot, label: "Identity", value: "Real user session" },
  { icon: BracketsCurly, label: "Contract", value: "JSON + schemas" },
  { icon: ShieldCheck, label: "Safety", value: "Explicit confirms" },
  { icon: Pulse, label: "Debugging", value: "Receipts + traces" },
];

export const OUTSIDE_IN_COMPARISON = {
  backend: {
    eyebrow: "YOUR BACKEND CAN PROVE",
    title: "The implementation responded.",
    items: [
      "The handler ran",
      "The payload matched the schema",
      "Telegram accepted the message",
      "The callback returned successfully",
    ],
  },
  experience: {
    eyebrow: "AGENT-TELEGRAM CAN VERIFY",
    title: "The experience actually worked.",
    items: [
      "The message appeared in the right chat",
      "The correct buttons were visible and usable",
      "The next state arrived in the expected order",
      "The complete flow worked from /start to the final reply",
    ],
  },
  artifacts: ["Message IDs", "Timings", "Actions", "Structured receipts"],
} as const;

export type UseCase = {
  icon: Icon;
  number: string;
  title: string;
  description: string;
  accent: string;
  tags: string[];
};

export const USE_CASES: UseCase[] = [
  {
    icon: CursorClick,
    number: "01",
    title: "Bot QA",
    description: "Run smoke, onboarding, keyboard, and regression flows against the bot users actually see.",
    accent: "cyan",
    tags: ["bot step", "press", "wait"],
  },
  {
    icon: ChatsCircle,
    number: "02",
    title: "Support agent",
    description: "Read context, search the inbox, draft a response, and keep the final send behind approval.",
    accent: "violet",
    tags: ["search", "draft", "send"],
  },
  {
    icon: Pulse,
    number: "03",
    title: "Community signal",
    description: "Watch selected groups and channels for questions, incidents, and conversations worth attention.",
    accent: "lime",
    tags: ["updates", "filter", "summary"],
  },
  {
    icon: PlugsConnected,
    number: "04",
    title: "Agent workflows",
    description: "Connect Telegram to CI, internal tools, and orchestration through CLI, JSON-RPC, or HTTP.",
    accent: "orange",
    tags: ["manifest", "OpenAPI", "HTTP"],
  },
];

export const SAFETY_POINTS: FeaturePoint[] = [
  { icon: Key, label: "Authentication", value: "QR sign-in stays local" },
  { icon: LockKey, label: "Session", value: "macOS Keychain support" },
  { icon: CheckCircle, label: "Actions", value: "Paid and destructive gates" },
  { icon: EyeSlash, label: "Observability", value: "Redacted logs by default" },
];

export const ARCHITECTURE_STEPS = [
  { icon: Robot, title: "AI agent", detail: "Codex, scripts, CI" },
  { icon: TerminalWindow, title: "CLI / HTTP", detail: "Schemas + receipts" },
  { icon: Pulse, title: "Local daemon", detail: "Unix socket IPC" },
  { icon: ChatsCircle, title: "Telegram", detail: "MTProto session" },
] as const;
