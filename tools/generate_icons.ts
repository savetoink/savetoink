/**
 * generate-icons.ts
 *
 * Converts a source SVG into all icons based on a CSV target list.
 *
 * Usage:
 *   bun generate-icons.ts <path-to-source.svg> <path-to-targets.csv>
 *
 * CSV format:
 *   type,path,size
 *   svg,cmd/webapp/src/lib/assets/favicon.svg
 *   png,cmd/extension/public/icon/16.png,16
 *
 * Note: 'size' is optional for SVG type
 */

import sharp from "sharp";
import fs from "fs/promises";
import path from "path";

// ── Types ─────────────────────────────────────────────────────────────────────

type IconType = "svg" | "png";

interface IconTarget {
  type: IconType;
  path: string;
  size?: number;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function showUsage(): void {
  console.error(`
Usage: bun generate-icons.ts <source.svg> <targets.csv>

Arguments:
  source.svg   Path to the source SVG file
  targets.csv  Path to CSV file containing icon targets

CSV format:
  type,path,size
  svg,cmd/webapp/src/lib/assets/favicon.svg
  png,cmd/extension/public/icon/16.png,16

  Note: 'size' is optional for SVG type
`);
}

function parseCsv(content: string): IconTarget[] {
  const lines = content.trim().split("\n");
  const targets: IconTarget[] = [];

  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;

    const parts = line.split(",");
    const type = parts[0] as IconType;
    const targetPath = parts[1];
    const size = parts[2] ? parseInt(parts[2], 10) : undefined;

    if (type === "svg") {
      targets.push({ type, path: targetPath });
    } else if (type === "png" && size !== undefined) {
      targets.push({ type, path: targetPath, size });
    }
  }

  return targets;
}

async function svgToPng(svgBuf: Buffer, size: number): Promise<Buffer> {
  return sharp(svgBuf)
    .resize(size, size, {
      fit: "contain",
      background: { r: 0, g: 0, b: 0, alpha: 0 },
    })
    .png()
    .toBuffer();
}

async function ensureDir(filePath: string): Promise<void> {
  const dir = path.dirname(filePath);
  await fs.mkdir(dir, { recursive: true });
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function main(): Promise<void> {
  const args = process.argv.slice(2);

  if (args.length < 2) {
    showUsage();
    process.exit(1);
  }

  const [srcSvg, csvPath] = args;

  console.log(`\n📂 Source : ${srcSvg}`);
  console.log(`📋 Targets: ${csvPath}\n`);

  const svgBuf: Buffer = await fs.readFile(srcSvg);
  const csvContent: string = await fs.readFile(csvPath, "utf-8");
  const targets: IconTarget[] = parseCsv(csvContent);

  const pngCache = new Map<number, Buffer>();

  for (const target of targets) {
    await ensureDir(target.path);

    if (target.type === "svg") {
      await fs.copyFile(srcSvg, target.path);
      console.log(`✅ ${target.path}`);
    } else {
      if (!pngCache.has(target.size!)) {
        pngCache.set(target.size!, await svgToPng(svgBuf, target.size!));
      }
      await fs.writeFile(target.path, pngCache.get(target.size!)!);
      console.log(`✅ ${target.path} (${target.size}×${target.size})`);
    }
  }

  console.log(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Done! ${targets.length} icons generated
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 `);
}

main().catch((err: Error) => {
  console.error("❌ Error:", err.message);
  process.exit(1);
});
