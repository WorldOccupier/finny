import type { CombinedTotal, Currency } from "../../api/dashboard";
import { formatMoney } from "./format";

interface NetWorthCardProps {
  totals: CombinedTotal[];
  currency: Currency;
  onCurrencyChange: (currency: Currency) => void;
}

export function NetWorthCard({ totals, currency, onCurrencyChange }: NetWorthCardProps) {
  const total = totals.find((item) => item.currency === currency)?.value ?? "0";

  return (
    <section className="net-worth-card">
      <div>
        <p className="eyebrow">Combined net worth</p>
        <h2>{formatMoney(total, currency)}</h2>
        <p className="muted">Across your UK and India assets</p>
      </div>
      <div className="currency-toggle" aria-label="Display currency">
        {(["GBP", "INR"] as Currency[]).map((option) => (
          <button
            aria-pressed={option === currency}
            className={option === currency ? "toggle-button active" : "toggle-button"}
            key={option}
            onClick={() => onCurrencyChange(option)}
            type="button"
          >
            {option}
          </button>
        ))}
      </div>
    </section>
  );
}
