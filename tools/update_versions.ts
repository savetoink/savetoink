import fs from "fs/promises";
import git from "isomorphic-git";
import * as jsonc from "jsonc-parser";
import path from "path";
import { getCurrentDevDate } from "@savetoink/shared";

type BumpType = "major" | "minor" | "patch" | "dev";

interface Version {
  major: number;
  minor: number;
  patch: number;
  devDate?: string;
  devCommitSha?: string;
}

interface IndentStyle {
  useTabs: boolean;
  spaceCount: number;
}

function showUsage(): void {
  console.error(`
Usage: bun update_versions.ts <major|minor|patch|dev>

Arguments:
  major  Increment major version, reset minor and patch to 0
  minor  Increment minor version, reset patch to 0
  patch  Increment patch version
  dev    Add or update dev suffix with date and commit SHA (MAJOR.MINOR.PATCH-dev+MMDDHHMM.COMMITSHA)

Example:
  bun update_versions.ts patch
  bun update_versions.ts dev
  `);
}

function parseVersion(versionStr: string): Version | null {
  const match = versionStr
    .trim()
    .match(/^(\d+)\.(\d+)\.(\d+)(?:-dev\+(\d{8})\.([a-f0-9]{7}))?$/);
  if (!match) return null;
  return {
    major: parseInt(match[1], 10),
    minor: parseInt(match[2], 10),
    patch: parseInt(match[3], 10),
    devDate: match[4] || undefined,
    devCommitSha: match[5] || undefined,
  };
}

function formatVersion(version: Version): string {
  if (version.devDate && version.devCommitSha) {
    return `${version.major}.${version.minor}.${version.patch}-dev+${version.devDate}.${version.devCommitSha}`;
  }
  return `${version.major}.${version.minor}.${version.patch}`;
}

async function getCurrentCommitSha(repoRoot: string): Promise<string> {
  const sha = await git.resolveRef({ fs, dir: repoRoot, ref: "HEAD" });
  return sha.slice(0, 7);
}

async function bumpVersion(
  version: Version,
  bumpType: BumpType,
  repoRoot: string,
): Promise<Version> {
  switch (bumpType) {
    case "major":
      return { major: version.major + 1, minor: 0, patch: 0 };
    case "minor":
      return { major: version.major, minor: version.minor + 1, patch: 0 };
    case "patch":
      return {
        major: version.major,
        minor: version.minor,
        patch: version.patch + 1,
      };
    case "dev": {
      const commitSha = await getCurrentCommitSha(repoRoot);
      const devDate = getCurrentDevDate();
      return {
        major: version.major,
        minor: version.minor,
        patch: version.patch,
        devDate,
        devCommitSha: commitSha,
      };
    }
    default:
      return version;
  }
}

async function detectIndent(filePath: string): Promise<IndentStyle> {
  const content = await fs.readFile(filePath, "utf-8");
  const lines = content.split("\n").slice(0, 10);

  let useTabs = false;
  let spaceCount = 2;
  let spaceCounts: number[] = [];

  for (const line of lines) {
    const match = line.match(/^(\s+)/);
    if (match) {
      const spaces = match[1];
      if (spaces.includes("\t")) {
        useTabs = true;
      } else if (spaces.length > 0 && /^\s*$/.test(spaces)) {
        spaceCounts.push(spaces.length);
      }
    }
  }

  if (!useTabs && spaceCounts.length > 0) {
    const counts = new Set(spaceCounts);
    if (counts.has(2)) spaceCount = 2;
    else if (counts.has(4)) spaceCount = 4;
    else spaceCount = Math.min(...counts);
  }

  return { useTabs, spaceCount };
}

function formatJson(obj: any, indentStyle: IndentStyle): string {
  const indent = indentStyle.useTabs
    ? "\t"
    : " ".repeat(indentStyle.spaceCount);
  return JSON.stringify(obj, null, indent);
}

async function updateBunLockWorkspaces(
  bunLockPath: string,
  repoRoot: string,
  newVersion: string,
): Promise<number> {
  const content = await fs.readFile(bunLockPath, "utf-8");

  let obj: any;

  try {
    obj = jsonc.parse(content);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    console.warn(
      `⚠️  Skipped ${path.relative(repoRoot, bunLockPath)} (invalid JSONC: ${message})`,
    );
    return 0;
  }

  if (!obj.workspaces || typeof obj.workspaces !== "object") {
    console.log(`✅ ${path.relative(repoRoot, bunLockPath)} (no workspaces)`);
    return 0;
  }

  const relativeBunLockPath = path.relative(repoRoot, bunLockPath);
  let updatedCount = 0;
  const replacements: { old: string; new: string }[] = [];

  for (const [workspaceName, workspaceData] of Object.entries(obj.workspaces)) {
    if (
      workspaceName === "" ||
      typeof workspaceData !== "object" ||
      !workspaceData
    )
      continue;
    if (!("version" in workspaceData)) continue;

    const oldVersion = workspaceData.version;
    if (oldVersion === newVersion) continue;

    replacements.push({
      old: `"version": "${oldVersion}"`,
      new: `"version": "${newVersion}"`,
    });

    console.log(
      `✅ ${relativeBunLockPath}:${workspaceName} (${oldVersion} → ${newVersion})`,
    );
    updatedCount++;
  }

  if (updatedCount === 0) {
    console.log(`✅ ${relativeBunLockPath} (no version fields to update)`);
    return 0;
  }

  let newContent = content;
  for (const { old, new: newVal } of replacements) {
    newContent = newContent.replace(old, newVal);
  }

  await fs.writeFile(bunLockPath, newContent);
  return updatedCount;
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);

  if (args.length < 1) {
    showUsage();
    process.exit(1);
  }

  const bumpType = args[0] as BumpType;
  if (
    bumpType !== "major" &&
    bumpType !== "minor" &&
    bumpType !== "patch" &&
    bumpType !== "dev"
  ) {
    console.error(
      "Error: argument must be 'major', 'minor', 'patch', or 'dev'",
    );
    showUsage();
    process.exit(1);
  }

  const repoRoot = path.resolve(process.cwd(), "..");
  const versionFilePath = path.join(repoRoot, "VERSION");

  const versionContent = await fs.readFile(versionFilePath, "utf-8");
  const version = parseVersion(versionContent);

  if (!version) {
    console.error("Error: VERSION file has invalid format");
    process.exit(1);
  }

  const newVersion = await bumpVersion(version, bumpType, repoRoot);
  const oldVersionStr = formatVersion(version);
  const newVersionStr = formatVersion(newVersion);

  const updateMessage =
    bumpType === "dev"
      ? `\n📦 Updating version: ${oldVersionStr} → ${newVersionStr} (dev suffix)\n`
      : `\n📦 Updating version: ${oldVersionStr} → ${newVersionStr} (${bumpType})\n`;
  console.log(updateMessage);

  const files = await git.listFiles({ fs, dir: repoRoot, ref: "HEAD" });
  const packageJsonFiles = files
    .filter((f) => f.endsWith("package.json"))
    .map((f) => path.join(repoRoot, f));

  await fs.writeFile(versionFilePath, newVersionStr + "\n");
  console.log("✅ Updated VERSION file");

  const goAppPath = path.join(repoRoot, "backend", "lib", "consts", "app.go");
  try {
    const goContent = await fs.readFile(goAppPath, "utf-8");
    const newGoContent = goContent.replace(
      /var version = ".*"/,
      `var version = "${newVersionStr}"`,
    );
    if (goContent !== newGoContent) {
      await fs.writeFile(goAppPath, newGoContent);
      console.log(`✅ ${path.relative(repoRoot, goAppPath)} (version updated)`);
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    console.warn(
      `⚠️  Could not update ${path.relative(repoRoot, goAppPath)}: ${message}`,
    );
  }

  const openapiPath = path.join(
    repoRoot,
    "backend",
    "lib",
    "internal",
    "server",
    "handlers",
    "openapi.yaml",
  );
  try {
    const yamlContent = await fs.readFile(openapiPath, "utf-8");
    const newYamlContent = yamlContent.replace(
      /version: \d+\.\d+\.\d+(?:-dev\+\d{8}\.[a-f0-9]{7})?/,
      `version: ${newVersionStr}`,
    );
    if (yamlContent !== newYamlContent) {
      await fs.writeFile(openapiPath, newYamlContent);
      console.log(
        `✅ ${path.relative(repoRoot, openapiPath)} (version updated)`,
      );
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    console.warn(
      `⚠️  Could not update ${path.relative(repoRoot, openapiPath)}: ${message}`,
    );
  }

  let updatedCount = 0;

  for (const filePath of packageJsonFiles) {
    const content = await fs.readFile(filePath, "utf-8");
    const indentStyle = await detectIndent(filePath);
    let obj: any;

    try {
      obj = JSON.parse(content);
    } catch {
      console.warn(
        `⚠️  Skipped ${path.relative(repoRoot, filePath)} (invalid JSON)`,
      );
      continue;
    }

    if (!obj.version) {
      console.log(`✅ ${path.relative(repoRoot, filePath)} (no version field)`);
      continue;
    }

    const oldVersion = obj.version;
    obj.version = newVersionStr;
    const newContent = formatJson(obj, indentStyle);

    if (content.endsWith("\n")) {
      await fs.writeFile(filePath, newContent + "\n");
    } else {
      await fs.writeFile(filePath, newContent);
    }

    console.log(
      `✅ ${path.relative(repoRoot, filePath)} (${oldVersion} → ${newVersionStr})`,
    );
    updatedCount++;
  }

  const bunLockPath = path.join(repoRoot, "frontend", "bun.lock");
  const packageCount = updatedCount;
  try {
    const bunLockUpdates = await updateBunLockWorkspaces(
      bunLockPath,
      repoRoot,
      newVersionStr,
    );
    updatedCount += bunLockUpdates;
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    console.warn(
      `⚠️  Could not update ${path.relative(repoRoot, bunLockPath)}: ${message}`,
    );
  }

  console.log(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Done! Updated VERSION, Go version constant, OpenAPI spec, ${packageCount} package.json files, and ${updatedCount - packageCount} bun.lock workspace versions
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  `);
}

main().catch((err: Error) => {
  console.error("❌ Error:", err.message);
  process.exit(1);
});
