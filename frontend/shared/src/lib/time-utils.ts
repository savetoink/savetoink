import { Temporal } from "./temporal";

export function formatInstantFromEpochMs(ms: number): string {
  const instant = Temporal.Instant.fromEpochMilliseconds(ms);
  const zonedDateTime = instant.toZonedDateTimeISO(Temporal.Now.timeZoneId());
  return zonedDateTime.toLocaleString("en-GB", {
    weekday: "short",
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    timeZoneName: "short",
  });
}

export function getCurrentDevDate(): string {
  const now = Temporal.Now.plainDateTimeISO();
  const day = String(now.day).padStart(2, "0");
  const month = String(now.month).padStart(2, "0");
  const hour = String(now.hour).padStart(2, "0");
  const minute = String(now.minute).padStart(2, "0");
  return `${day}${month}${hour}${minute}`;
}

export function getCurrentYear(): number {
  return Temporal.Now.plainDateTimeISO().year;
}

export function formatTimeRemainingFromEpochMs(ms: number): string {
  const now = Temporal.Now.instant();
  const target = Temporal.Instant.fromEpochMilliseconds(ms);
  const duration = target.since(now);

  if (duration.sign < 0) {
    return "0h 0m 0s";
  }

  const totalSeconds = duration.seconds;
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = Math.floor(totalSeconds % 60);

  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds}s`;
  } else {
    return `${seconds}s`;
  }
}
