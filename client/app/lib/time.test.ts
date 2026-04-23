import { describe, it, expect } from "vitest";

import { parseRelativeTime, formatRelativeTime } from "./time";

describe("parseRelativeTime", () => {
  it("parses minutes", () => {
    expect(parseRelativeTime("30m")).toBe(30 * 60_000);
  });

  it("parses hours", () => {
    expect(parseRelativeTime("6h")).toBe(6 * 60 * 60_000);
  });

  it("parses days", () => {
    expect(parseRelativeTime("2d")).toBe(2 * 24 * 60 * 60_000);
  });

  it("parses months", () => {
    expect(parseRelativeTime("1mo")).toBe(30 * 24 * 60 * 60_000);
  });

  it("parses compound h+m", () => {
    expect(parseRelativeTime("6h30m")).toBe(6 * 60 * 60_000 + 30 * 60_000);
  });

  it("parses compound d+h+m", () => {
    expect(parseRelativeTime("1d12h30m")).toBe(
      24 * 60 * 60_000 + 12 * 60 * 60_000 + 30 * 60_000,
    );
  });

  it("parses compound mo+d", () => {
    expect(parseRelativeTime("2mo3d")).toBe(
      2 * 30 * 24 * 60 * 60_000 + 3 * 24 * 60 * 60_000,
    );
  });

  it("trims whitespace", () => {
    expect(parseRelativeTime("  3h  ")).toBe(3 * 60 * 60_000);
  });

  it("allows interior whitespace", () => {
    expect(parseRelativeTime("6h   30m")).toBe(6 * 60 * 60_000 + 30 * 60_000);
  });

  it("returns null for empty string", () => {
    expect(parseRelativeTime("")).toBeNull();
  });

  it("returns null for bare number", () => {
    expect(parseRelativeTime("30")).toBeNull();
  });

  it("returns null for unknown unit", () => {
    expect(parseRelativeTime("5y")).toBeNull();
  });

  it("returns null for garbage between tokens", () => {
    expect(parseRelativeTime("6hfoo30m")).toBeNull();
  });

  it("returns null for trailing garbage", () => {
    expect(parseRelativeTime("6h30mxyz")).toBeNull();
  });
});

describe("formatRelativeTime", () => {
  function tsSecondsAgo(s: number): string {
    return new Date(Date.now() - s * 1000).toISOString();
  }

  it("formats seconds ago", () => {
    expect(formatRelativeTime(tsSecondsAgo(30))).toBe("30s ago");
  });

  it("formats minutes ago", () => {
    expect(formatRelativeTime(tsSecondsAgo(5 * 60))).toBe("5m ago");
  });

  it("formats hours ago", () => {
    expect(formatRelativeTime(tsSecondsAgo(3 * 60 * 60))).toBe("3h ago");
  });

  it("formats days ago", () => {
    expect(formatRelativeTime(tsSecondsAgo(2 * 24 * 60 * 60))).toBe("2d ago");
  });

  it("uses 'now' for very recent timestamps", () => {
    expect(formatRelativeTime(tsSecondsAgo(0))).toBe("now");
  });
});
