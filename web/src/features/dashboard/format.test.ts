import { describe, expect, it } from "vitest";
import { chartTicks, formatMoney, scaledChartValue } from "./format";

describe("dashboard formatting", () => {
  it("formats decimal strings without converting through floating point", () => {
    expect(formatMoney("12345678901234567890.125", "GBP")).toBe("£12,345,678,901,234,567,890.12");
    expect(formatMoney("50000", "INR")).toBe("₹50,000.00");
  });

  it("scales chart values with integer arithmetic", () => {
    expect(scaledChartValue("1000.50")).toBe(100050n);
  });

  it("creates decimal-safe chart ticks from the minimum to maximum value", () => {
    expect(chartTicks(["1000", "1500"])).toEqual(["1500.00", "1333.33", "1166.66", "1000.00"]);
  });
});
