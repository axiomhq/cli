#!/usr/bin/env node
"use strict";

// Runs the Axiom CLI binary from the platform package that npm installed as
// an optional dependency of the axiom package.

const { spawnSync } = require("node:child_process");

const pkg = `@axiomhq/cli-${process.platform}-${process.arch}`;
const exe = process.platform === "win32" ? "axiom.exe" : "axiom";

let bin;
try {
  bin = require.resolve(`${pkg}/bin/${exe}`);
} catch {
  console.error(
    `The Axiom CLI does not ship a binary for ${process.platform}-${process.arch}, ` +
      `or npm did not install ${pkg} (optional dependencies may be disabled).\n` +
      "Reinstall without --omit=optional, or see https://github.com/axiomhq/cli#usage for other ways to install it.",
  );
  process.exit(1);
}

const result = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
if (result.signal) {
  process.kill(process.pid, result.signal);
}
process.exit(result.status);
