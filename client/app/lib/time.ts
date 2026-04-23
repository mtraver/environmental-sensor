import { useEffect, useState } from "react";

const RTF = new Intl.RelativeTimeFormat("en", {
  numeric: "auto",
  style: "narrow",
});

const UNIT_MS: Record<string, number> = {
  mo: 30 * 24 * 60 * 60_000,
  d: 24 * 60 * 60_000,
  h: 60 * 60_000,
  m: 60_000,
};

export function parseRelativeTime(input: string): number | null {
  const FULL = /^(\d+mo|\d+d|\d+h|\d+m)(\s*(\d+mo|\d+d|\d+h|\d+m))*$/;
  const TOKEN = /(\d+)(mo|d|h|m)/g;
  const trimmed = input.trim();

  if (!FULL.test(trimmed)) return null;

  let total = 0;
  for (const match of trimmed.matchAll(TOKEN)) {
    total += parseInt(match[1], 10) * UNIT_MS[match[2]];
  }
  return total;
}

export function formatRelativeTime(timestamp: string): string {
  const diffMs = Date.now() - new Date(timestamp).getTime();
  const diffSec = Math.floor(diffMs / 1000);

  if (diffSec < 60) {
    return RTF.format(-diffSec, "second");
  }

  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) {
    return RTF.format(-diffMin, "minute");
  }

  const diffHours = Math.floor(diffMin / 60);
  if (diffHours < 24) {
    return RTF.format(-diffHours, "hour");
  }

  const diffDays = Math.floor(diffHours / 24);
  return RTF.format(-diffDays, "day");
}

export function useRelativeTime(timestamp: string): string {
  const [, setTick] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 15_000);
    return () => clearInterval(id);
  }, [timestamp]);

  return formatRelativeTime(timestamp);
}
