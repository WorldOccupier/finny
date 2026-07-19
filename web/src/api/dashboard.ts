export type ValueType = "UKGBP" | "INDIAINR";
export type Currency = "GBP" | "INR";

export interface AssetValue {
  type: ValueType;
  value: string;
}

export interface Asset {
  id: number;
  name: string;
  values: AssetValue[];
  valueTypes?: ValueType[];
}

export interface TotalValue {
  type: ValueType;
  value: string;
}

export interface CombinedTotal {
  currency: Currency;
  value: string;
}

export interface DashboardTotals {
  country: TotalValue[];
  combined: CombinedTotal[];
}

export interface Snapshot {
  id: number;
  committedAt: string;
  fxRate: string;
  assets: Asset[];
  totals: DashboardTotals;
}

export interface SpendingLimit {
  key: string;
  amount: string;
  currency: Currency;
}

export interface IncomeTotals {
  userOneGBP: string;
  userTwoGBP: string;
}

export interface DashboardResponse {
  revision: number;
  assets: Asset[];
  currentFxRate: string;
  currentTotals: DashboardTotals;
  history: Snapshot[];
  spendingLimits: SpendingLimit[];
  income: IncomeTotals;
}

export interface DashboardRequest {
  revision: number;
  assets: Asset[];
  fxRate: string;
  spendingLimits: SpendingLimit[];
  income: IncomeTotals;
}

export class DashboardApiError extends Error {
  readonly status?: number;
  readonly code?: string;

  constructor(message: string, status?: number, code?: string) {
    super(message);
    this.name = "DashboardApiError";
    this.status = status;
    this.code = code;
  }
}

const DECIMAL_PATTERN = /^\d+(?:\.\d+)?$/;
const LONDON_TIMESTAMP_PATTERN =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})$/;

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

function isDecimalString(value: unknown): value is string {
  return typeof value === "string" && DECIMAL_PATTERN.test(value);
}

function isZeroDecimal(value: string): boolean {
  const [whole, fraction] = value.split(".");
  return /^0+$/.test(whole) && (fraction === undefined || /^0+$/.test(fraction));
}

function isValueType(value: unknown): value is ValueType {
  return value === "UKGBP" || value === "INDIAINR";
}

function isCurrency(value: unknown): value is Currency {
  return value === "GBP" || value === "INR";
}

function isAssetValue(value: unknown): value is AssetValue {
  return isRecord(value) && isValueType(value.type) && isDecimalString(value.value);
}

function isAsset(value: unknown): value is Asset {
  return (
    isRecord(value) &&
    isNonNegativeInteger(value.id) &&
    typeof value.name === "string" &&
    value.name.trim().length > 0 &&
    Array.isArray(value.values) &&
    value.values.length > 0 &&
    value.values.every(isAssetValue) &&
    (value.valueTypes === undefined || (Array.isArray(value.valueTypes) && value.valueTypes.every(isValueType)))
  );
}

function isTotalValue(value: unknown): value is TotalValue {
  return isRecord(value) && isValueType(value.type) && isDecimalString(value.value);
}

function isCombinedTotal(value: unknown): value is CombinedTotal {
  return isRecord(value) && isCurrency(value.currency) && isDecimalString(value.value);
}

function hasExactlyOneOfEach<T>(values: T[], matches: ((value: T) => boolean)[]): boolean {
  return values.length === matches.length && matches.every((match) => values.filter(match).length === 1);
}

function hasCompleteTotals(country: unknown[], combined: unknown[]): boolean {
  return (
    hasExactlyOneOfEach(country, [
      (total) => isTotalValue(total) && total.type === "UKGBP",
      (total) => isTotalValue(total) && total.type === "INDIAINR",
    ]) &&
    hasExactlyOneOfEach(combined, [
      (total) => isCombinedTotal(total) && total.currency === "GBP",
      (total) => isCombinedTotal(total) && total.currency === "INR",
    ])
  );
}

function isDashboardTotals(value: unknown, allowEmpty: boolean): value is DashboardTotals {
  return (
    isRecord(value) &&
    Array.isArray(value.country) &&
    value.country.every(isTotalValue) &&
    Array.isArray(value.combined) &&
    value.combined.every(isCombinedTotal) &&
    ((allowEmpty && value.country.length === 0 && value.combined.length === 0) || hasCompleteTotals(value.country, value.combined))
  );
}

function isSnapshot(value: unknown): value is Snapshot {
  return (
    isRecord(value) &&
    isNonNegativeInteger(value.id) &&
    typeof value.committedAt === "string" &&
    LONDON_TIMESTAMP_PATTERN.test(value.committedAt) &&
    !Number.isNaN(Date.parse(value.committedAt)) &&
    isDecimalString(value.fxRate) &&
    Array.isArray(value.assets) &&
    value.assets.every(isAsset) &&
    isDashboardTotals(value.totals, false)
  );
}

function isSpendingLimit(value: unknown): value is SpendingLimit {
  return (
    isRecord(value) &&
    typeof value.key === "string" &&
    value.key.trim().length > 0 &&
    isDecimalString(value.amount) &&
    isCurrency(value.currency)
  );
}

function isIncomeTotals(value: unknown): value is IncomeTotals {
  return (
    isRecord(value) &&
    isDecimalString(value.userOneGBP) &&
    isDecimalString(value.userTwoGBP)
  );
}

export function isDashboardResponse(value: unknown): value is DashboardResponse {
  if (
    isRecord(value) &&
    isNonNegativeInteger(value.revision) &&
    Array.isArray(value.assets) &&
    value.assets.every(isAsset) &&
    isDecimalString(value.currentFxRate) &&
    Array.isArray(value.history) &&
    value.history.every(isSnapshot) &&
    Array.isArray(value.spendingLimits) &&
    value.spendingLimits.every(isSpendingLimit) &&
    isIncomeTotals(value.income)
  ) {
    const isEmpty =
      value.revision === 0 &&
      value.assets.length === 0 &&
      value.history.length === 0 &&
      value.spendingLimits.length === 0 &&
      isZeroDecimal(value.income.userOneGBP) &&
      isZeroDecimal(value.income.userTwoGBP) &&
      isZeroDecimal(value.currentFxRate);
    return isDashboardTotals(value.currentTotals, isEmpty);
  }
  return false;
}

export async function fetchDashboard(signal?: AbortSignal): Promise<DashboardResponse> {
  const response = await fetch("/api/dashboard", { signal });
  if (!response.ok) {
    throw new DashboardApiError("We could not load your dashboard right now.");
  }

  try {
    const data: unknown = await response.json();
    if (!isDashboardResponse(data)) {
      throw new DashboardApiError("The dashboard response was not valid.");
    }
    return data;
  } catch (error) {
    if (error instanceof DashboardApiError) throw error;
    throw new DashboardApiError("The dashboard response was not valid.");
  }
}

export function createIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export async function saveDashboard(
  request: DashboardRequest,
  idempotencyKey = createIdempotencyKey(),
): Promise<DashboardResponse> {
  let response: Response;
  try {
    response = await fetch("/api/dashboard", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": idempotencyKey,
      },
      body: JSON.stringify(request),
    });
  } catch {
    throw new DashboardApiError("We could not save your dashboard right now.");
  }

  let body: unknown;
  try {
    body = await response.json();
  } catch {
    throw new DashboardApiError("The server returned an invalid response.", response.status);
  }

  if (!response.ok) {
    const error = isRecord(body) && isRecord(body.error) ? body.error : undefined;
    throw new DashboardApiError(
      typeof error?.message === "string" ? error.message : "We could not save your dashboard right now.",
      response.status,
      typeof error?.code === "string" ? error.code : undefined,
    );
  }
  if (!isDashboardResponse(body)) {
    throw new DashboardApiError("The saved dashboard response was not valid.", response.status);
  }
  return body;
}
