import { describe, expect, it } from "vitest";

import {
  AGENT_SETUP_PROMPT,
  BOT_SCENARIOS,
  GITHUB_URL,
  INSTALL_COMMAND,
  OUTSIDE_IN_COMPARISON,
  SAFETY_POINTS,
  SKILL_INSTALL_COMMAND,
  USE_CASES,
} from "./content";

describe("agent-telegram landing content", () => {
  it("leads with end-to-end bot testing", () => {
    expect(BOT_SCENARIOS[0].id).toBe("onboarding");
    expect(BOT_SCENARIOS[0].command).toContain("bot step");
    expect(USE_CASES[0].title).toBe("Bot QA");
  });

  it("explains outside-in evidence without a fake conversation", () => {
    expect(OUTSIDE_IN_COMPARISON.backend.items).toContain("The handler ran");
    expect(OUTSIDE_IN_COMPARISON.experience.items).toContain(
      "The complete flow worked from /start to the final reply",
    );
    expect(OUTSIDE_IN_COMPARISON.artifacts).toContain("Structured receipts");
  });

  it("uses real install and repository targets", () => {
    expect(INSTALL_COMMAND).toBe("npm install -g agent-telegram");
    expect(SKILL_INSTALL_COMMAND).toBe("agent-telegram skills install agent-telegram");
    expect(GITHUB_URL).toBe("https://github.com/dukaev/agent-telegram");
  });

  it("leads with a safe prompt-first CLI and skill setup", () => {
    expect(AGENT_SETUP_PROMPT).toContain(INSTALL_COMMAND);
    expect(AGENT_SETUP_PROMPT).toContain(SKILL_INSTALL_COMMAND);
    expect(AGENT_SETUP_PROMPT).not.toMatch(/QR code/i);
    expect(AGENT_SETUP_PROMPT).toMatch(/read-only/i);
    expect(AGENT_SETUP_PROMPT).toMatch(/Ask before sending messages/i);
  });

  it("keeps scenarios uniquely addressable", () => {
    const ids = BOT_SCENARIOS.map((scenario) => scenario.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it("makes local auth and confirmation gates visible", () => {
    const safetyCopy = SAFETY_POINTS.map((point) => `${point.label} ${point.value}`).join(" ");
    const scenarioCopy = BOT_SCENARIOS.map((scenario) => scenario.assertion).join(" ");

    expect(safetyCopy).toMatch(/local/i);
    expect(scenarioCopy).toMatch(/confirmation required/i);
  });
});
