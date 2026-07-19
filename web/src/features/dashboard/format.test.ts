import { describe, expect, it } from "vitest";
import { chartTicks, combinedTotalValue, formatMoney, scaledChartValue } from "./format";

describe("dashboard formatting", () => {
  it("formats decimal strings without converting through floating point", () => {
    expect(formatMoney("12345678901234567890.125", "GBP")).toBe("£12,345,678,901,234,567,890.12");
    expect(formatMoney("50000", "INR")).toBe("₹50,000.00");
  });

  it("selects the actual combined value for the requested currency", () => {
    const totals = [
      { currency: "GBP" as const, value: "125.50" },
      { currency: "INR" as const, value: "17570" },
    ];

    expect(combinedTotalValue(totals, "GBP")).toBe("125.50");
    expect(combinedTotalValue(totals, "INR")).toBe("17570");
    expect(combinedTotalValue(totals, "GBP")).not.toBe(combinedTotalValue(totals, "INR"));
  });

  it("scales chart values with integer arithmetic", () => {
    expect(scaledChartValue("1000.50")).toBe(100050n);
  });

  it("creates decimal-safe chart ticks from the minimum to maximum value", () => {
    expect(chartTicks(["1000", "1500"])).toEqual(["1500.00", "1333.33", "1166.66", "1000.00"]);
  });
});
