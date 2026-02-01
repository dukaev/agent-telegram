#!/usr/bin/env node

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const zlib = require("zlib");

const PACKAGE = require("../package.json");
const BINARY_NAME = "agent-telegram";
const REPO = "user/agent-telegram"; // Change to your GitHub username/repo

// Map Node.js platform/arch to Go GOOS/GOARCH
const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
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
  // GoReleaser archive format: agent-telegram_0.1.0_darwin_arm64.tar.gz
  const ext = platform === "windows" ? "zip" : "tar.gz";
  return `https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}_${version}_${platform}_${arch}.${ext}`;
}

function fetch(url) {
  return new Promise((resolve, reject) => {
    const request = (url) => {
      https
        .get(url, (response) => {
          if (response.statusCode === 302 || response.statusCode === 301) {
            request(response.headers.location);
            return;
          }
          if (response.statusCode !== 200) {
            reject(new Error(`HTTP ${response.statusCode}: ${url}`));
            return;
          }
          resolve(response);
        })
        .on("error", reject);
    };
    request(url);
  });
}

async function downloadAndExtract(url, destDir, binaryName) {
  console.log(`Downloading from ${url}...`);

  const response = await fetch(url);
  const chunks = [];

  await new Promise((resolve, reject) => {
    response.on("data", (chunk) => chunks.push(chunk));
    response.on("end", resolve);
    response.on("error", reject);
  });

  const buffer = Buffer.concat(chunks);

  if (url.endsWith(".tar.gz")) {
    // Extract tar.gz using tar command
    const archivePath = path.join(destDir, "archive.tar.gz");
    fs.writeFileSync(archivePath, buffer);
    execSync(`tar -xzf "${archivePath}" -C "${destDir}"`, { stdio: "pipe" });
    fs.unlinkSync(archivePath);
  } else if (url.endsWith(".zip")) {
    // For Windows zip files
    const archivePath = path.join(destDir, "archive.zip");
    fs.writeFileSync(archivePath, buffer);
    try {
      execSync(`unzip -o "${archivePath}" -d "${destDir}"`, { stdio: "pipe" });
    } catch {
      // Try PowerShell on Windows
      execSync(
        `powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${destDir}' -Force"`,
        { stdio: "pipe" }
      );
    }
    fs.unlinkSync(archivePath);
  }

  // Find and move binary to bin directory
  const extractedBinary = findBinary(destDir, binaryName);
  if (extractedBinary) {
    const finalPath = path.join(destDir, binaryName);
    if (extractedBinary !== finalPath) {
      fs.renameSync(extractedBinary, finalPath);
    }
    fs.chmodSync(finalPath, 0o755);
  }

  // Clean up extra files from archive
  const keepFiles = [binaryName, "run.js"];
  for (const file of fs.readdirSync(destDir)) {
    if (!keepFiles.includes(file)) {
      const filePath = path.join(destDir, file);
      const stat = fs.statSync(filePath);
      if (stat.isFile() && !file.endsWith(".js")) {
        // Keep JS files, remove others like LICENSE, README
        fs.unlinkSync(filePath);
      }
    }
  }
}

function findBinary(dir, binaryName) {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    const filePath = path.join(dir, file);
    const stat = fs.statSync(filePath);
    if (stat.isDirectory()) {
      const found = findBinary(filePath, binaryName);
      if (found) return found;
    } else if (file === binaryName) {
      return filePath;
    }
  }
  return null;
}

async function main() {
  const binDir = path.join(__dirname, "..", "bin");
  const { platform, arch } = getPlatform();
  const binaryName = getBinaryName(platform);
  const binaryPath = path.join(binDir, binaryName);

  // Skip if binary already exists
  if (fs.existsSync(binaryPath)) {
    console.log(`Binary already exists at ${binaryPath}`);
    return;
  }

  const url = getDownloadUrl(PACKAGE.version, platform, arch);

  try {
    await downloadAndExtract(url, binDir, binaryName);
    console.log(`Successfully installed ${BINARY_NAME} v${PACKAGE.version}`);
  } catch (error) {
    console.error(`Failed to download binary: ${error.message}`);
    console.error("");
    console.error("Manual installation options:");
    console.error(`  1. Download from: https://github.com/${REPO}/releases`);
    console.error("  2. Build from source: go build -o bin/agent-telegram .");
    process.exit(1);
  }
}

main();
