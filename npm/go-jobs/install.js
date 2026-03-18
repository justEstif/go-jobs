"use strict";
const https = require("https");
const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const os = require("os");

const BINARY = "jobs";
const REPO = "justEstif/go-jobs";
const VERSION = require("./package.json").version;

const PLATFORM_MAP = {
  "linux-x64":    { os: "linux",   arch: "amd64", ext: ".tar.gz" },
  "linux-arm64":  { os: "linux",   arch: "arm64", ext: ".tar.gz" },
  "darwin-x64":   { os: "darwin",  arch: "amd64", ext: ".tar.gz" },
  "darwin-arm64": { os: "darwin",  arch: "arm64", ext: ".tar.gz" },
  "win32-x64":    { os: "windows", arch: "amd64", ext: ".zip"    },
};

const key = `${process.platform}-${process.arch}`;
const plat = PLATFORM_MAP[key];
if (!plat) {
  console.error(`@justestif/go-jobs: unsupported platform "${key}"`);
  console.error(`Supported: ${Object.keys(PLATFORM_MAP).join(", ")}`);
  process.exit(1);
}

const exe = process.platform === "win32" ? `${BINARY}.exe` : BINARY;
const archiveName = `jobs_v${VERSION}_${plat.os}_${plat.arch}${plat.ext}`;
const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
const binDir = path.join(__dirname, "bin");
const tmpFile = path.join(os.tmpdir(), archiveName);
const outPath = path.join(binDir, exe);

if (fs.existsSync(outPath)) {
  process.exit(0); // already installed
}

fs.mkdirSync(binDir, { recursive: true });

console.log(`Downloading jobs from ${url}...`);

function tryUnlink(p) {
  try { fs.unlinkSync(p); } catch {}
}

function download(url, dest, cb) {
  const file = fs.createWriteStream(dest);
  https.get(url, (res) => {
    if (res.statusCode === 301 || res.statusCode === 302) {
      file.close();
      tryUnlink(dest);
      return download(res.headers.location, dest, cb);
    }
    if (res.statusCode !== 200) {
      file.close();
      tryUnlink(dest);
      return cb(new Error(`HTTP ${res.statusCode} for ${url}`));
    }
    res.pipe(file);
    file.on("finish", () => file.close(cb));
  }).on("error", (err) => {
    tryUnlink(dest);
    cb(err);
  });
}

download(url, tmpFile, (err) => {
  if (err) {
    console.error(`Failed to download jobs: ${err.message}`);
    process.exit(1);
  }

  try {
    if (plat.ext === ".zip") {
      execFileSync("unzip", ["-j", tmpFile, exe, "-d", binDir]);
    } else {
      execFileSync("tar", ["-xzf", tmpFile, "--strip-components=0", "-C", binDir]);
    }
    fs.chmodSync(outPath, 0o755);
    fs.unlinkSync(tmpFile);
    console.log(`Installed jobs to ${outPath}`);
  } catch (e) {
    console.error(`Failed to extract jobs: ${e.message}`);
    process.exit(1);
  }
});
