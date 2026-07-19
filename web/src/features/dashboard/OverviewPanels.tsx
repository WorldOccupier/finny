import type { IncomeTotals, SpendingLimit } from "../../api/dashboard";
import { formatMoney } from "./format";

export function SpendingLimits({ limits }: { limits: SpendingLimit[] }) {
  return (
    <section className="panel detail-panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Monthly guardrails</p>
          <h2>Spending limits</h2>
        </div>
      </div>
      {limits.length === 0 ? (
        <p className="muted">No spending limits configured.</p>
      ) : (
        limits.map((limit) => (
          <div className="detail-row" key={limit.key}>
            <span>{limit.key}</span>
            <strong>{formatMoney(limit.amount, limit.currency)}</strong>
          </div>
        ))
      )}
    </section>
  );
}

export function Income({ income }: { income: IncomeTotals }) {
  return (
    <section className="panel detail-panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Household</p>
          <h2>Monthly income</h2>
        </div>
        <span className="country-badge">GBP</span>
      </div>
      <div className="detail-row">
        <span>User one</span>
        <strong>{formatMoney(income.userOneGBP, "GBP")}</strong>
      </div>
      <div className="detail-row">
        <span>User two</span>
        <strong>{formatMoney(income.userTwoGBP, "GBP")}</strong>
      </div>
    </section>
  );
}
