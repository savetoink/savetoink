/**
 * generate-icons.ts
 *
 * Converts a source SVG into all icons needed for:
 *   - Extension  (16, 32, 48, 96, 128 px PNGs)
 *   - Webapp     (favicon.svg)
 *   - Website    (favicon.svg)
 *
 * Usage:
 *   npx tsx generate-icons.ts [path-to-source.svg]
 *
 * Defaults:
 *   source  → ./icon.svg
 */

import sharp from "sharp";
import fs from "fs/promises";
import path from "path";

// ── Config ────────────────────────────────────────────────────────────────────

const srcSvg: string = process.argv[2] ?? "./favicon.svg";

const ICON_TARGETS = [
  {
    type: "svg" as const,
    path: "cmd/webapp/src/lib/assets/favicon.svg",
  },
  {
    type: "svg" as const,
    path: "cmd/website/public/favicon.svg",
  },
  {
    type: "png" as const,
    path: "cmd/extension/public/icon/16.png",
    size: 16,
  },
  {
    type: "png" as const,
    path: "cmd/extension/public/icon/32.png",
    size: 32,
  },
  {
    type: "png" as const,
    path: "cmd/extension/public/icon/48.png",
    size: 48,
  },
  {
    type: "png" as const,
    path: "cmd/extension/public/icon/96.png",
    size: 96,
  },
  {
    type: "png" as const,
    path: "cmd/extension/public/icon/128.png",
    size: 128,
  },
];

// ── Helpers ───────────────────────────────────────────────────────────────────

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
  console.log(`\n📂 Source : ${srcSvg}\n`);

  const svgBuf: Buffer = await fs.readFile(srcSvg);

  const pngCache = new Map<number, Buffer>();

  for (const target of ICON_TARGETS) {
    await ensureDir(target.path);

    if (target.type === "svg") {
      await fs.copyFile(srcSvg, target.path);
      console.log(`✅ ${target.path}`);
    } else {
      if (!pngCache.has(target.size)) {
        pngCache.set(target.size, await svgToPng(svgBuf, target.size));
      }
      await fs.writeFile(target.path, pngCache.get(target.size)!);
      console.log(`✅ ${target.path} (${target.size}×${target.size})`);
    }
  }

  console.log(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Done! ${ICON_TARGETS.length} icons generated
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`);
}

main().catch((err: Error) => {
  console.error("❌ Error:", err.message);
  process.exit(1);
});
