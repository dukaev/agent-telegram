#!/usr/bin/env node

const { spawn } = require("child_process");
const path = require("path");

const BINARY_NAME = process.platform === "win32" ? "agent-telegram.exe" : "agent-telegram";
const binaryPath = path.join(__dirname, BINARY_NAME);

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  env: process.env,
});

child.on("error", (err) => {
  if (err.code === "ENOENT") {
    console.error(`Binary not found at ${binaryPath}`);
    console.error("Run 'npm run postinstall' or build from source");
    process.exit(1);
  }
  throw err;
});

child.on("exit", (code) => {
  process.exit(code ?? 0);
});
