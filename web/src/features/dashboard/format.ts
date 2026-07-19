import type { Currency, ValueType } from "../../api/dashboard";

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
