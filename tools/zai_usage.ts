import { Temporal } from "@js-temporal/polyfill";

const ZAI_API_KEY = process.env.ZAI_API_KEY;

if (!ZAI_API_KEY) {
  console.error("Error: ZAI_API_KEY environment variable is not set");
  process.exit(1);
}

async function main(): Promise<void> {
  const response = await fetch(
    "https://api.z.ai/api/monitor/usage/quota/limit",
    {
      headers: {
        Authorization: `Bearer ${ZAI_API_KEY}`,
      },
    },
  );

  if (!response.ok) {
    console.error(`Error: API request failed with status ${response.status}`);
    process.exit(1);
  }

  const data = (await response.json()) as {
    data: {
      limits: Array<{
        type: string;
        percentage: number;
        nextResetTime?: number;
      }>;
    };
  };
  const tokensLimit = data.data.limits.find(
    (limit) => limit.type === "TOKENS_LIMIT",
  );

  if (!tokensLimit) {
    console.error("Error: TOKENS_LIMIT not found in response");
    process.exit(1);
  }

  console.log(`z.ai Token Usage: ${tokensLimit.percentage}%`);
  const resetDate: Temporal.Instant | null = tokensLimit.nextResetTime
    ? Temporal.Instant.fromEpochMilliseconds(tokensLimit.nextResetTime)
    : null;
  if (resetDate) {
    console.log(`Resets at: ${resetDate.toLocaleString()}`);
  }
}

main().catch((err: Error) => {
  console.error("Error:", err.message);
  process.exit(1);
});
