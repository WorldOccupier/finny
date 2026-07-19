import type { CombinedTotal, Currency, ValueType } from "../../api/dashboard";

export const currencyForType = (type: ValueType): Currency =>
  type === "UKGBP" ? "GBP" : "INR";

export const countryForType = (type: ValueType): string =>
  type === "UKGBP" ? "United Kingdom" : "India";

export function formatMoney(value: string, currency: Currency): string {
  const [whole = "0", fraction = ""] = value.split(".");
  const sign = whole.startsWith("-") ? "-" : "";
  const unsignedWhole = sign ? whole.slice(1) : whole;
  const grouped = unsignedWhole.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  const decimals = fraction.padEnd(2, "0").slice(0, 2);
  const symbol = currency === "GBP" ? "£" : "₹";
  return `${sign}${symbol}${grouped}.${decimals}`;
}

export function combinedTotalValue(totals: CombinedTotal[], currency: Currency): string | undefined {
  return totals.find((total) => total.currency === currency)?.value;
}

export function formatDate(timestamp: string): string {
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return "Unknown date";
  return new Intl.DateTimeFormat("en-GB", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    timeZone: "Europe/London",
  }).format(date);
}

export function scaledChartValue(value: string): bigint {
  const [whole = "0", fraction = ""] = value.split(".");
  const sign = whole.startsWith("-") ? -1n : 1n;
  const unsignedWhole = sign === -1n ? whole.slice(1) : whole;
  const cents = BigInt(`${unsignedWhole || "0"}${fraction.padEnd(2, "0").slice(0, 2)}`);
  return cents * sign;
}

function decimalFromScaledValue(value: bigint): string {
  const sign = value < 0n ? "-" : "";
  const absolute = value < 0n ? -value : value;
  const whole = absolute / 100n;
  const fraction = (absolute % 100n).toString().padStart(2, "0");
  return `${sign}${whole}.${fraction}`;
}

export function chartTicks(values: string[]): string[] {
  const scaledValues = values.map(scaledChartValue);
  const minimum = scaledValues.reduce((current, value) => (value < current ? value : current));
  const maximum = scaledValues.reduce((current, value) => (value > current ? value : current));
  const range = maximum - minimum;
  return [3n, 2n, 1n, 0n].map((position) =>
    decimalFromScaledValue(minimum + (range * position) / 3n),
  );
}
