#!/usr/bin/env bun
/**
 * sonar-report.ts
 * Queries SonarQube Cloud and outputs a Markdown report for feeding to a coding agent.
 *
 * Usage:
 *   SONAR_TOKEN=<token> SONAR_PROJECT=<org_repo> npx tsx sonar-report.ts
 *   SONAR_TOKEN=<token> SONAR_PROJECT=<org_repo> bun sonar-report.ts
 *   SONAR_TOKEN=<token> SONAR_PROJECT=<org_repo> bun sonar-report.ts <pr_number>
 *
 * Optional env vars:
 *   SONAR_MAX_ISSUES=<n>     — max issues to fetch (default: 20)
 *   SONAR_BASE_URL=<url>     — defaults to https://sonarcloud.io
 */

const TOKEN = process.env.SONAR_TOKEN;
const PROJECT = process.env.SONAR_PROJECT;
const PR = process.argv[2];
const MAX_ISSUES = parseInt(process.env.SONAR_MAX_ISSUES ?? "20", 10);
const BASE_URL = process.env.SONAR_BASE_URL ?? "https://sonarcloud.io";

if (!TOKEN) {
  console.error("Error: SONAR_TOKEN env var is required");
  process.exit(1);
}
if (!PROJECT) {
  console.error("Error: SONAR_PROJECT env var is required");
  process.exit(1);
}

// --- Types ---

interface QualityGateCondition {
  metricKey: string;
  actualValue: string;
  status: "OK" | "ERROR" | "WARN";
  comparator: string;
  errorThreshold?: string;
}

interface QualityGateStatus {
  status: "OK" | "ERROR" | "WARN" | "NONE";
  conditions: QualityGateCondition[];
}

interface Measure {
  metric: string;
  value: string;
  bestValue?: boolean;
}

interface Issue {
  key: string;
  rule: string;
  severity: "BLOCKER" | "CRITICAL" | "MAJOR" | "MINOR" | "INFO";
  message: string;
  component: string;
  line?: number;
  type: "BUG" | "VULNERABILITY" | "CODE_SMELL";
  effort?: string;
  tags: string[];
}

// --- API helpers ---

async function get<T>(
  path: string,
  params: Record<string, string> = {},
): Promise<T> {
  const url = new URL(`${BASE_URL}${path}`);
  for (const [k, v] of Object.entries(params)) url.searchParams.set(k, v);

  const res = await fetch(url.toString(), {
    headers: { Authorization: `Bearer ${TOKEN}` },
  });

  if (!res.ok) {
    const body = await res.text();
    throw new Error(`SonarQube API error ${res.status} on ${path}: ${body}`);
  }

  return res.json() as Promise<T>;
}

// --- Report sections ---

async function fetchQualityGate(): Promise<QualityGateStatus> {
  const params: Record<string, string> = { projectKey: PROJECT! };
  if (PR) params.pullRequest = PR;

  const data = await get<{ projectStatus: QualityGateStatus }>(
    "/api/qualitygates/project_status",
    params,
  );
  return data.projectStatus;
}

async function fetchMetrics(): Promise<Measure[]> {
  const metrics = [
    "bugs",
    "vulnerabilities",
    "code_smells",
    "coverage",
    "duplicated_lines_density",
    "ncloc",
    "reliability_rating",
    "security_rating",
    "sqale_rating",
  ].join(",");

  const params: Record<string, string> = {
    component: PROJECT!,
    metricKeys: metrics,
  };
  if (PR) params.pullRequest = PR;

  const data = await get<{ component: { measures: Measure[] } }>(
    "/api/measures/component",
    params,
  );
  return data.component.measures;
}

async function fetchIssues(): Promise<{ issues: Issue[]; total: number }> {
  const params: Record<string, string> = {
    projectKeys: PROJECT!,
    resolved: "false",
    ps: String(MAX_ISSUES),
    s: "SEVERITY",
    asc: "false",
  };
  if (PR) params.pullRequest = PR;

  const data = await get<{ issues: Issue[]; total: number }>(
    "/api/issues/search",
    params,
  );
  return { issues: data.issues, total: data.total };
}

// --- Formatting helpers ---

const RATING_LABELS: Record<string, string> = {
  "1.0": "A",
  "2.0": "B",
  "3.0": "C",
  "4.0": "D",
  "5.0": "E",
};

const METRIC_LABELS: Record<string, string> = {
  bugs: "Bugs",
  vulnerabilities: "Vulnerabilities",
  code_smells: "Code Smells",
  coverage: "Coverage (%)",
  duplicated_lines_density: "Duplication (%)",
  ncloc: "Lines of Code",
  reliability_rating: "Reliability",
  security_rating: "Security",
  sqale_rating: "Maintainability",
};

function formatMetricValue(metric: string, value: string): string {
  if (metric.endsWith("_rating")) return RATING_LABELS[value] ?? value;
  if (metric === "coverage" || metric === "duplicated_lines_density")
    return `${value}%`;
  return value;
}

function statusEmoji(status: string): string {
  return status === "OK" ? "✅" : status === "ERROR" ? "❌" : "⚠️";
}

function severityEmoji(severity: string): string {
  const map: Record<string, string> = {
    BLOCKER: "🔴",
    CRITICAL: "🟠",
    MAJOR: "🟡",
    MINOR: "🔵",
    INFO: "⚪",
  };
  return map[severity] ?? "•";
}

function shortComponent(component: string): string {
  // Strip the project prefix (e.g. "myorg_myrepo:internal/foo.go" → "internal/foo.go")
  return component.includes(":")
    ? component.split(":").slice(1).join(":")
    : component;
}

// --- Main ---

async function main() {
  const timestamp = new Date().toISOString();
  const scope = PR ? `PR #${PR}` : "main branch";

  const [qualityGate, measures, { issues, total }] = await Promise.all([
    fetchQualityGate(),
    fetchMetrics(),
    fetchIssues(),
  ]);

  const lines: string[] = [];

  lines.push(`# SonarQube Cloud Report`);
  lines.push(
    `**Project:** \`${PROJECT}\` | **Scope:** ${scope} | **Generated:** ${timestamp}`,
  );
  lines.push("");

  // Quality Gate
  lines.push(
    `## ${statusEmoji(qualityGate.status)} Quality Gate: ${qualityGate.status}`,
  );
  if (qualityGate.conditions.length > 0) {
    lines.push("");
    lines.push("| Metric | Value | Threshold | Status |");
    lines.push("|--------|-------|-----------|--------|");
    for (const c of qualityGate.conditions) {
      const threshold = c.errorThreshold
        ? `${c.comparator} ${c.errorThreshold}`
        : "—";
      lines.push(
        `| ${c.metricKey} | ${c.actualValue} | ${threshold} | ${statusEmoji(c.status)} ${c.status} |`,
      );
    }
  }
  lines.push("");

  // Metrics summary
  lines.push("## 📊 Metrics");
  lines.push("");
  lines.push("| Metric | Value |");
  lines.push("|--------|-------|");
  for (const m of measures) {
    const label = METRIC_LABELS[m.metric] ?? m.metric;
    const value = formatMetricValue(m.metric, m.value);
    lines.push(`| ${label} | ${value} |`);
  }
  lines.push("");

  // Issues
  const shown = Math.min(issues.length, MAX_ISSUES);
  lines.push(`## 🐛 Issues (showing ${shown} of ${total})`);
  lines.push("");

  if (issues.length === 0) {
    lines.push("No open issues found. 🎉");
  } else {
    lines.push("| Severity | Type | File | Line | Message | Effort |");
    lines.push("|----------|------|------|------|---------|--------|");
    for (const issue of issues) {
      const file = shortComponent(issue.component);
      const line = issue.line ?? "—";
      const effort = issue.effort ?? "—";
      const msg = issue.message.replace(/\|/g, "\\|");
      lines.push(
        `| ${severityEmoji(issue.severity)} ${issue.severity} | ${issue.type} | \`${file}\` | ${line} | ${msg} | ${effort} |`,
      );
    }
  }

  lines.push("");
  lines.push("---");
  lines.push(
    `_Report generated by sonar-report.ts. Full dashboard: ${BASE_URL}/project/overview?id=${PROJECT}_`,
  );

  console.log(lines.join("\n"));
}

main().catch((err) => {
  console.error(err.message);
  process.exit(1);
});
