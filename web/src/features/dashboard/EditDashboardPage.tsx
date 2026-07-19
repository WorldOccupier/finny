import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import {
  DashboardApiError,
  createIdempotencyKey,
  fetchDashboard,
  saveDashboard,
  type Asset,
  type DashboardRequest,
  type DashboardResponse,
  type SpendingLimit,
} from "../../api/dashboard";

type FormState = Pick<DashboardRequest, "revision" | "assets" | "fxRate" | "spendingLimits" | "income">;
type SaveAttempt = { fingerprint: string; key: string };

function toForm(data: DashboardResponse): FormState {
  return {
    revision: data.revision,
    fxRate: data.currentFxRate === "0" ? "1" : data.currentFxRate,
    assets: data.assets.map((asset) => ({
      ...asset,
      values: [
        { type: "UKGBP" as const, value: asset.values.find((value) => value.type === "UKGBP")?.value ?? "0" },
        { type: "INDIAINR" as const, value: asset.values.find((value) => value.type === "INDIAINR")?.value ?? "0" },
      ],
    })),
    spendingLimits: data.spendingLimits.map((limit) => ({ ...limit })),
    income: { ...data.income },
  };
}

export function EditDashboardPage() {
  const [form, setForm] = useState<FormState | null>(null);
  const [status, setStatus] = useState<"loading" | "ready" | "saving">("loading");
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<"error" | "success" | "info">("info");
  const saveAttempt = useRef<SaveAttempt | undefined>(undefined);

  const load = () => {
    setStatus("loading");
    setMessage(undefined);
    fetchDashboard()
      .then((data) => {
        setForm(toForm(data));
        setStatus("ready");
      })
      .catch((error: unknown) => {
        setForm(null);
        setStatus("ready");
        setMessage(error instanceof DashboardApiError ? error.message : "We could not load the editor.");
        setMessageKind("error");
      });
  };

  useEffect(load, []);

  if (status === "loading" || form === null) {
    return (
      <main className="status-shell">
        <div className="status-mark">✦</div>
        <p className="eyebrow">finny</p>
        <h1>{status === "loading" ? "Loading editor" : "Editor unavailable"}</h1>
        <p>{message ?? "Gathering the latest dashboard data…"}</p>
        {message && <button className="primary-button" onClick={load} type="button">Try again</button>}
      </main>
    );
  }

  const updateAsset = (id: number, update: (asset: Asset) => Asset) => {
    setForm((current) => current && { ...current, assets: current.assets.map((asset) => asset.id === id ? update(asset) : asset) });
  };

  const addAsset = () => {
    const nextID = form.assets.reduce((highest, asset) => Math.max(highest, asset.id), -1) + 1;
    setForm({
      ...form,
      assets: [...form.assets, { id: nextID, name: "New asset", values: [{ type: "UKGBP", value: "0" }, { type: "INDIAINR", value: "0" }] }],
    });
  };

  const updateLimit = (index: number, update: (limit: SpendingLimit) => SpendingLimit) => {
    setForm({ ...form, spendingLimits: form.spendingLimits.map((limit, current) => current === index ? update(limit) : limit) });
  };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setStatus("saving");
    setMessage(undefined);
    const fingerprint = JSON.stringify(form);
    const attempt = saveAttempt.current?.fingerprint === fingerprint
      ? saveAttempt.current
      : { fingerprint, key: createIdempotencyKey() };
    saveAttempt.current = attempt;
    try {
      const committed = await saveDashboard(form, attempt.key);
      setForm(toForm(committed));
      saveAttempt.current = undefined;
      setStatus("ready");
      setMessage("Your snapshot is saved.");
      setMessageKind("success");
    } catch (error: unknown) {
      setStatus("ready");
      if (error instanceof DashboardApiError && error.status === 409 && error.code === "revision_conflict") {
        setMessage("This dashboard changed elsewhere. Reload the latest values before saving.");
        setMessageKind("error");
      } else {
        setMessage(error instanceof DashboardApiError ? error.message : "We could not save your dashboard.");
        setMessageKind("error");
      }
    }
  };

  return (
    <main className="app-shell editor-shell">
      <header className="topbar">
        <a className="brand" href="/">finny<span>.</span></a>
        <a className="edit-link" href="/">Back to dashboard <span aria-hidden="true">↗</span></a>
      </header>
      <section className="hero editor-hero">
        <div>
          <p className="eyebrow">Shape your snapshot</p>
          <h1>Edit dashboard.</h1>
          <p className="hero-copy">Keep your household picture current, one calm update at a time.</p>
        </div>
        <span className="revision">Revision {form.revision}</span>
      </section>
      {message && <div className={`editor-message ${messageKind}`} role="status">{message}</div>}
      {messageKind === "error" && message?.includes("changed elsewhere") && <button className="secondary-button" onClick={load} type="button">Reload latest dashboard</button>}
      <form onSubmit={submit}>
        <section className="panel form-panel">
          <div className="section-heading">
            <div><p className="eyebrow">Shared portfolio</p><h2>Assets</h2></div>
            <button className="secondary-button" onClick={addAsset} type="button">+ Add asset</button>
          </div>
          {form.assets.length === 0 && <p className="empty-panel">Your first asset can be added here. Give it both a UK and India value.</p>}
          <div className="editor-list">
            {form.assets.map((asset) => (
              <fieldset className="asset-editor" key={asset.id}>
                <legend>Asset {asset.id + 1}</legend>
                <div className="asset-editor-heading">
                  <label>Asset name<input aria-label="Asset name" required value={asset.name} onChange={(event) => updateAsset(asset.id, (current) => ({ ...current, name: event.target.value }))} /></label>
                  <button className="text-button danger" onClick={() => setForm({ ...form, assets: form.assets.filter((current) => current.id !== asset.id) })} type="button">Remove</button>
                </div>
                <div className="value-grid">
                  <label>United Kingdom · GBP<input inputMode="decimal" aria-label={`${asset.name} UK GBP value`} required value={asset.values.find((value) => value.type === "UKGBP")?.value ?? "0"} onChange={(event) => updateAsset(asset.id, (current) => ({ ...current, values: current.values.map((value) => value.type === "UKGBP" ? { ...value, value: event.target.value } : value) }))} /></label>
                  <label>India · INR<input inputMode="decimal" aria-label={`${asset.name} India INR value`} required value={asset.values.find((value) => value.type === "INDIAINR")?.value ?? "0"} onChange={(event) => updateAsset(asset.id, (current) => ({ ...current, values: current.values.map((value) => value.type === "INDIAINR" ? { ...value, value: event.target.value } : value) }))} /></label>
                </div>
              </fieldset>
            ))}
          </div>
        </section>
        <div className="two-column editor-grid">
          <section className="panel form-panel">
            <div className="section-heading"><div><p className="eyebrow">Conversion</p><h2>Exchange rate</h2></div><span className="country-badge">INR / GBP</span></div>
            <label>Indian rupees per pound<input inputMode="decimal" aria-label="Indian rupees per pound" required value={form.fxRate} onChange={(event) => setForm({ ...form, fxRate: event.target.value })} /></label>
            <p className="field-hint">This rate is frozen into the snapshot’s history.</p>
          </section>
          <section className="panel form-panel">
            <div className="section-heading"><div><p className="eyebrow">Household</p><h2>Monthly income</h2></div><span className="country-badge">GBP</span></div>
            <div className="value-grid">
              <label>User one<input inputMode="decimal" aria-label="User one income" required value={form.income.userOneGBP} onChange={(event) => setForm({ ...form, income: { ...form.income, userOneGBP: event.target.value } })} /></label>
              <label>User two<input inputMode="decimal" aria-label="User two income" required value={form.income.userTwoGBP} onChange={(event) => setForm({ ...form, income: { ...form.income, userTwoGBP: event.target.value } })} /></label>
            </div>
          </section>
        </div>
        <section className="panel form-panel">
          <div className="section-heading"><div><p className="eyebrow">Monthly guardrails</p><h2>Spending limits</h2></div><button className="secondary-button" onClick={() => setForm({ ...form, spendingLimits: [...form.spendingLimits, { key: "New limit", amount: "0", currency: "GBP" }] })} type="button">+ Add limit</button></div>
          {form.spendingLimits.length === 0 && <p className="empty-panel">No limits yet. Add one to keep a little space around your spending.</p>}
          <div className="editor-list">
            {form.spendingLimits.map((limit, index) => <div className="limit-editor" key={`${limit.key}-${index}`}>
              <label>Limit name<input aria-label={`Spending limit ${index + 1} name`} required value={limit.key} onChange={(event) => updateLimit(index, (current) => ({ ...current, key: event.target.value }))} /></label>
              <label>Amount<input inputMode="decimal" aria-label={`Spending limit ${index + 1} amount`} required value={limit.amount} onChange={(event) => updateLimit(index, (current) => ({ ...current, amount: event.target.value }))} /></label>
              <label>Currency<select aria-label={`Spending limit ${index + 1} currency`} value={limit.currency} onChange={(event) => updateLimit(index, (current) => ({ ...current, currency: event.target.value as "GBP" | "INR" }))}><option>GBP</option><option>INR</option></select></label>
              <button className="text-button danger" onClick={() => setForm({ ...form, spendingLimits: form.spendingLimits.filter((_, current) => current !== index) })} type="button">Remove</button>
            </div>)}
          </div>
        </section>
        <div className="form-actions"><a className="secondary-button" href="/">Cancel</a><button className="primary-button" disabled={status === "saving"} type="submit">{status === "saving" ? "Saving…" : "Save snapshot"}</button></div>
      </form>
      <footer>Finny · private by design</footer>
    </main>
  );
}
