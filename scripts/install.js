#!/usr/bin/env node

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const PACKAGE = require("../package.json");
const BINARY_NAME = "agent-telegram";

// Map Node.js platform/arch to Go GOOS/GOARCH
const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
  arm: "arm",
};

function getPlatform() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    throw new Error(
      `Unsupported platform: ${process.platform}-${process.arch}`
    );
  }

  return { platform, arch };
}

function getBinaryName(platform) {
  return platform === "windows" ? `${BINARY_NAME}.exe` : BINARY_NAME;
}

function getDownloadUrl(version, platform, arch) {
  const ext = platform === "windows" ? ".exe" : "";
  // Adjust this URL pattern to match your GitHub releases
  // Example: https://github.com/user/agent-telegram/releases/download/v0.1.0/agent-telegram-darwin-arm64
  return `https://github.com/user/agent-telegram/releases/download/v${version}/${BINARY_NAME}-${platform}-${arch}${ext}`;
}

async function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    const request = (url) => {
      https
        .get(url, (response) => {
          if (response.statusCode === 302 || response.statusCode === 301) {
            // Follow redirect
            request(response.headers.location);
            return;
          }

          if (response.statusCode !== 200) {
            reject(new Error(`Download failed: ${response.statusCode} ${url}`));
            return;
          }

          response.pipe(file);
          file.on("finish", () => {
            file.close();
            resolve();
          });
        })
        .on("error", (err) => {
          fs.unlink(dest, () => {});
          reject(err);
        });
    };

    request(url);
  });
}

async function main() {
  const binDir = path.join(__dirname, "..", "bin");
  const { platform, arch } = getPlatform();
  const binaryName = getBinaryName(platform);
  const binaryPath = path.join(binDir, binaryName);

  // Skip if binary already exists (local dev)
  if (fs.existsSync(binaryPath)) {
    console.log(`Binary already exists at ${binaryPath}`);
    return;
  }

  const url = getDownloadUrl(PACKAGE.version, platform, arch);
  console.log(`Downloading ${BINARY_NAME} from ${url}...`);

  try {
    await download(url, binaryPath);

    // Make executable on Unix
    if (platform !== "windows") {
      fs.chmodSync(binaryPath, 0o755);
    }

    console.log(`Successfully installed ${BINARY_NAME}`);
  } catch (error) {
    console.error(`Failed to download binary: ${error.message}`);
    console.error(
      "You may need to build from source: go build -o bin/agent-telegram ."
    );
    process.exit(1);
  }
}

main();
