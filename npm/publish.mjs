#!/usr/bin/env node

// Publishes the binaries GoReleaser built into dist/ to npm: one package per
// platform (@axiomhq/cli-<platform>-<arch>) holding the binary, and the axiom
// package that depends on all of them optionally and runs the one npm installed
// for the current platform.
//
// Until v1.0.0, the axiom package was the Axiom AI SDK (now @axiomhq/ai), so
// CLI releases below v1.0.1 can't be published to it.
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
    description: `The ${os}-${cpu} binary of the Axiom CLI (axiom).`,
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
  name: "axiom",
  ...common,
  description: "Powerful log analytics from the comfort of your command-line",
  bin: { axiom: "bin/axiom.js" },
  engines: { node: ">=18" },
  optionalDependencies: Object.fromEntries(packages.map((p) => [p.name, version])),
});

// Platform packages go first so axiom never points at missing versions.
const all = [...packages, { name: "axiom", dir: main }].map((p) => ({
  ...p,
  published: dryRun ? undefined : publishedRepository(p.name),
}));

// Check every package before publishing any, so a version taken by another
// project fails the release without leaving some packages published.
for (const { name, published } of all) {
  if (published !== undefined && published !== common.repository.url) {
    throw new Error(`${name}@${version} is already taken by ${published || "another project"}`);
  }
}

for (const { name, dir, published } of all) {
  if (published !== undefined) {
    console.log(`${name}@${version} is already published, skipping`);
  } else {
    publish(dir);
  }
}

function publish(dir) {
  const args = ["publish", "--access", "public", "--tag", version.includes("-") ? "next" : "latest"];
  if (dryRun) args.push("--dry-run");
  else if (process.env.GITHUB_ACTIONS) args.push("--provenance");
  execFileSync("npm", args, { cwd: dir, stdio: "inherit" });
}

// Returns the repository URL of name@version, "" if it has none, or undefined
// if that version isn't published.
function publishedRepository(name) {
  let out;
  try {
    out = execFileSync("npm", ["view", `${name}@${version}`, "repository.url"], { encoding: "utf8", stdio: "pipe" });
  } catch {
    return undefined;
  }
  return out.trim();
}

function writeJSON(path, value) {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`);
}
