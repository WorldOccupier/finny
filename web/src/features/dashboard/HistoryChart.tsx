import type { CombinedTotal, Snapshot, Currency } from "../../api/dashboard";
import { formatDate, formatMoney, scaledChartValue } from "./format";

interface HistoryChartProps {
  history: Snapshot[];
  currency: Currency;
}

export function HistoryChart({ history, currency }: HistoryChartProps) {
  const points = history
    .map((snapshot) => ({
      date: snapshot.committedAt,
      total: snapshot.totals.combined.find((item) => item.currency === currency),
    }))
    .filter((point): point is { date: string; total: CombinedTotal } => point.total !== undefined);

  if (points.length === 0) {
    return <section className="panel empty-panel">Your net-worth history will appear here after your first snapshot.</section>;
  }

  const values = points.map((point) => scaledChartValue(point.total.value));
  const minimum = values.reduce((current, value) => (value < current ? value : current));
  const maximum = values.reduce((current, value) => (value > current ? value : current));
  const range = maximum - minimum || 1n;
  const chartPoints = points.map((point, index) => {
    const normalized = Number(((values[index] - minimum) * 10000n) / range) / 10000;
    return `${(index / Math.max(points.length - 1, 1)) * 100},${88 - normalized * 62}`;
  });

  return (
    <section className="panel history-panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Momentum</p>
          <h2>Net-worth history</h2>
        </div>
        <span className="chart-caption">{points.length} snapshot{points.length === 1 ? "" : "s"}</span>
      </div>
      <div className="chart" role="img" aria-label={`Net worth history in ${currency}`}>
        <svg viewBox="0 0 100 100" preserveAspectRatio="none">
          <defs>
            <linearGradient id="chart-fill" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor="#8f7cff" stopOpacity=".32" />
              <stop offset="100%" stopColor="#8f7cff" stopOpacity="0" />
            </linearGradient>
          </defs>
          <polygon points={`0,88 ${chartPoints.join(" ")} 100,88`} fill="url(#chart-fill)" />
          <polyline points={chartPoints.join(" ")} fill="none" stroke="#8f7cff" strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.4" vectorEffect="non-scaling-stroke" />
        </svg>
      </div>
      <div className="chart-labels">
        <span>{formatDate(points[0].date)}</span>
        <strong>{formatMoney(points[points.length - 1].total.value, currency)}</strong>
        <span>{formatDate(points[points.length - 1].date)}</span>
      </div>
    </section>
  );
}
