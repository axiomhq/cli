#!/usr/bin/env node

// Publishes the binaries GoReleaser built into dist/ to npm: one package per
// platform (@axiomhq/cli-<platform>-<arch>) holding the binary, and the
// @axiomhq/cli package that depends on all of them optionally and runs the one
// npm installed for the current platform.
//
// Usage: node npm/publish.mjs [--dry-run]

import { execFileSync } from "node:child_process";
import { chmodSync, copyFileSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const dryRun = process.argv.includes("--dry-run");
const root = join(import.meta.dirname, "..");
const dist = join(root, "dist");
const out = join(dist, "npm");

// Snapshot builds report versions like "v0.22.0-next".
const version = JSON.parse(readFileSync(join(dist, "metadata.json"), "utf8")).version.replace(/^v/, "");
const artifacts = JSON.parse(readFileSync(join(dist, "artifacts.json"), "utf8"));

const platforms = { darwin: "darwin", linux: "linux", windows: "win32" };
const archs = { amd64: "x64", arm64: "arm64", 386: "ia32", arm: "arm" };

const common = {
  version,
  license: "MIT",
  homepage: "https://axiom.co",
  repository: { type: "git", url: "git+https://github.com/axiomhq/cli.git" },
};

rmSync(out, { recursive: true, force: true });

const binaries = artifacts.filter(
  (a) =>
    a.type === "Binary" &&
    a.extra.ID !== "docker" &&
    // Node.js supports 32-bit ARM from ARMv7 on, so skip the ARMv6 build.
    (a.goarch !== "arm" || a.goarm === "7"),
);

const packages = binaries.map((a) => {
  const os = platforms[a.goos];
  const cpu = archs[a.goarch];
  const name = `@axiomhq/cli-${os}-${cpu}`;
  const dir = join(out, `cli-${os}-${cpu}`);
  const exe = `axiom${a.extra.Ext}`;

  mkdirSync(join(dir, "bin"), { recursive: true });
  copyFileSync(join(root, a.path), join(dir, "bin", exe));
  chmodSync(join(dir, "bin", exe), 0o755);
  copyFileSync(join(root, "LICENSE"), join(dir, "LICENSE"));
  writeJSON(join(dir, "package.json"), {
    name,
    ...common,
    description: `The ${os}-${cpu} binary of the Axiom CLI (@axiomhq/cli).`,
    os: [os],
    cpu: [cpu],
    preferUnplugged: true,
  });
  return { name, dir };
});

const main = join(out, "cli");
mkdirSync(join(main, "bin"), { recursive: true });
copyFileSync(join(root, "npm", "axiom.js"), join(main, "bin", "axiom.js"));
copyFileSync(join(root, "README.md"), join(main, "README.md"));
copyFileSync(join(root, "LICENSE"), join(main, "LICENSE"));
writeJSON(join(main, "package.json"), {
  name: "@axiomhq/cli",
  ...common,
  description: "Powerful log analytics from the comfort of your command-line",
  bin: { axiom: "bin/axiom.js" },
  engines: { node: ">=18" },
  optionalDependencies: Object.fromEntries(packages.map((p) => [p.name, version])),
});

// Platform packages go first so @axiomhq/cli never points at missing versions.
for (const { name, dir } of [...packages, { name: "@axiomhq/cli", dir: main }]) {
  publish(name, dir);
}

function publish(name, dir) {
  if (!dryRun && isPublished(name)) {
    console.log(`${name}@${version} is already published, skipping`);
    return;
  }
  const args = ["publish", "--access", "public", "--tag", version.includes("-") ? "next" : "latest"];
  if (dryRun) args.push("--dry-run");
  else if (process.env.GITHUB_ACTIONS) args.push("--provenance");
  execFileSync("npm", args, { cwd: dir, stdio: "inherit" });
}

function isPublished(name) {
  try {
    execFileSync("npm", ["view", `${name}@${version}`, "version"], { stdio: "pipe" });
    return true;
  } catch {
    return false;
  }
}

function writeJSON(path, value) {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`);
}
