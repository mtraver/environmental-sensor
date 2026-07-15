import { describe, it, expect } from "vitest";

import { METRIC_ORDER } from "./metrics";

describe("METRIC_ORDER", () => {
  it("elements are unique", () => {
    expect(new Set(METRIC_ORDER).size).toBe(METRIC_ORDER.length);
  });
});
