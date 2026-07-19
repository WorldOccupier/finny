import { useEffect, useState } from "react";
import { DashboardApiError, fetchDashboard, type Currency, type DashboardResponse } from "../../api/dashboard";
import { AssetSection } from "./AssetSection";
import { HistoryChart } from "./HistoryChart";
import { Income, SpendingLimits } from "./OverviewPanels";
import { NetWorthCard } from "./NetWorthCard";

type LoadState = { status: "loading" } | { status: "error"; message: string } | { status: "ready"; data: DashboardResponse };

export function DashboardPage() {
  const [state, setState] = useState<LoadState>({ status: "loading" });
  const [currency, setCurrency] = useState<Currency>("GBP");

  useEffect(() => {
    const controller = new AbortController();
    fetchDashboard(controller.signal)
      .then((data) => setState({ status: "ready", data }))
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === "AbortError") return;
        const message = error instanceof DashboardApiError ? error.message : "Something went wrong while loading your dashboard.";
        setState({ status: "error", message });
      });
    return () => controller.abort();
  }, []);

  if (state.status === "loading") {
    return <StatusScreen title="Loading your dashboard" detail="Gathering your latest numbers…" />;
  }

  if (state.status === "error") {
    return (
      <StatusScreen
        title="Your dashboard is taking a moment"
        detail={state.message}
        action="Try again"
        onAction={() => window.location.reload()}
      />
    );
  }

  const { data } = state;
  const isEmpty = data.assets.length === 0 && data.history.length === 0;
  if (isEmpty) {
    return (
      <StatusScreen
        title="Your financial picture starts here"
        detail="Add your first snapshot to see your net worth, assets, and history in one place."
        action="Open editor"
        onAction={() => {
          window.location.href = "/edit";
        }}
      />
    );
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <a className="brand" href="/">
          finny<span>.</span>
        </a>
        <a className="edit-link" href="/edit">
          Edit dashboard <span aria-hidden="true">↗</span>
        </a>
      </header>
      <section className="hero">
        <div>
          <p className="eyebrow">Dashboard</p>
          <h1>Your financial picture.</h1>
          <p className="hero-copy">A calm view of where your money is working for you.</p>
        </div>
        <span className="revision">Updated snapshot {data.revision}</span>
      </section>
      <NetWorthCard totals={data.currentTotals.combined} currency={currency} onCurrencyChange={setCurrency} />
      <HistoryChart history={data.history} currency={currency} />
      <div className="two-column">
        <AssetSection assets={data.assets} type="UKGBP" />
        <AssetSection assets={data.assets} type="INDIAINR" />
      </div>
      <div className="two-column">
        <SpendingLimits limits={data.spendingLimits} />
        <Income income={data.income} />
      </div>
      <footer>Finny · private by design</footer>
    </main>
  );
}

function StatusScreen({ title, detail, action, onAction }: { title: string; detail: string; action?: string; onAction?: () => void }) {
  return (
    <main className="status-shell">
      <div className="status-mark">✦</div>
      <p className="eyebrow">finny</p>
      <h1>{title}</h1>
      <p>{detail}</p>
      {action && (
        <button className="primary-button" onClick={onAction} type="button">
          {action}
        </button>
      )}
    </main>
  );
}
