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
type EditableSpendingLimit = SpendingLimit & { localID: string };
type EditableFormState = Omit<FormState, "spendingLimits"> & { spendingLimits: EditableSpendingLimit[] };
type SaveAttempt = { fingerprint: string; key: string };
type AssetCurrencySelection = "GBP" | "INR" | "BOTH";
type AssetValueType = "UKGBP" | "INDIAINR";

let nextLimitID = 0;

function newLimitID(): string {
  nextLimitID += 1;
  return `limit-${nextLimitID}`;
}

function toForm(data: DashboardResponse): EditableFormState {
  return {
    revision: data.revision,
    fxRate: data.currentFxRate === "0" ? "1" : data.currentFxRate,
    assets: data.assets.map((asset) => ({
      ...asset,
      values: asset.values.map((value) => ({ ...value })),
      valueTypes: asset.valueTypes ?? asset.values.map((value) => value.type),
    })),
    spendingLimits: data.spendingLimits.map((limit) => ({ ...limit, localID: newLimitID() })),
    income: { ...data.income },
  };
}

function toRequest(form: EditableFormState): FormState {
  return {
    ...form,
    spendingLimits: form.spendingLimits.map(({ localID: _localID, ...limit }) => limit),
  };
}

function selectedCurrencies(asset: Asset): AssetValueType[] {
  return asset.valueTypes ?? asset.values.map((value) => value.type);
}

function currencySelection(asset: Asset): AssetCurrencySelection {
  const currencies = selectedCurrencies(asset);
  const hasGBP = currencies.includes("UKGBP");
  const hasINR = currencies.includes("INDIAINR");
  if (hasGBP && hasINR) return "BOTH";
  return hasINR ? "INR" : "GBP";
}

function duplicateAssetIDs(assets: Asset[]): Set<number> {
  const counts = new Map<string, number>();
  for (const asset of assets) {
    const name = asset.name.trim().toLocaleLowerCase();
    counts.set(name, (counts.get(name) ?? 0) + 1);
  }
  return new Set(assets.filter((asset) => (counts.get(asset.name.trim().toLocaleLowerCase()) ?? 0) > 1).map((asset) => asset.id));
}

export function EditDashboardPage() {
  const [form, setForm] = useState<EditableFormState | null>(null);
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
      assets: [...form.assets, { id: nextID, name: "New asset", valueTypes: ["UKGBP"], values: [{ type: "UKGBP", value: "" }] }],
    });
  };

  const updateLimit = (localID: string, update: (limit: EditableSpendingLimit) => EditableSpendingLimit) => {
    setForm((current) => current && { ...current, spendingLimits: current.spendingLimits.map((limit) => limit.localID === localID ? update(limit) : limit) });
  };

  const updateAssetValue = (assetID: number, type: "UKGBP" | "INDIAINR", value: string) => {
    updateAsset(assetID, (asset) => ({
      ...asset,
      valueTypes: asset.valueTypes?.includes(type) ? asset.valueTypes : [...(asset.valueTypes ?? []), type],
      values: asset.values.some((item) => item.type === type)
        ? asset.values.map((item) => item.type === type ? { ...item, value } : item)
        : [...asset.values, { type, value }],
    }));
  };

  const updateAssetCurrencies = (assetID: number, selection: AssetCurrencySelection) => {
    const types: AssetValueType[] = selection === "GBP" ? ["UKGBP"] : selection === "INR" ? ["INDIAINR"] : ["UKGBP", "INDIAINR"];
    updateAsset(assetID, (asset) => ({
      ...asset,
      valueTypes: types,
      values: types.map((type) => asset.values.find((item) => item.type === type) ?? { type, value: "" }),
    }));
  };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (duplicateAssetIDs(form.assets).size > 0) {
      setMessage("Asset names must be unique, ignoring capitalisation and surrounding spaces.");
      setMessageKind("error");
      return;
    }
    setStatus("saving");
    setMessage(undefined);
    const request = toRequest(form);
    const fingerprint = JSON.stringify(request);
    const attempt = saveAttempt.current?.fingerprint === fingerprint
      ? saveAttempt.current
      : { fingerprint, key: createIdempotencyKey() };
    saveAttempt.current = attempt;
    try {
      const committed = await saveDashboard(request, attempt.key);
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
          {form.assets.length === 0 && <p className="empty-panel">Add an asset in GBP, INR, or both currencies.</p>}
          <div className="editor-list">
            {form.assets.map((asset) => (
              <fieldset className="asset-editor" key={asset.id}>
                <legend>Asset {asset.id + 1}</legend>
                <div className="asset-editor-heading">
                  <label>Asset name<input aria-label={`Asset ${asset.id + 1} name`} aria-describedby="duplicate-asset-names" aria-invalid={duplicateAssetIDs(form.assets).has(asset.id)} required value={asset.name} onChange={(event) => updateAsset(asset.id, (current) => ({ ...current, name: event.target.value }))} /></label>
                  <button className="icon-button danger" aria-label={`Remove ${asset.name}`} title={`Remove ${asset.name}`} onClick={() => setForm({ ...form, assets: form.assets.filter((current) => current.id !== asset.id) })} type="button"><TrashIcon /></button>
                </div>
                <AssetCurrencyFields asset={asset} selection={currencySelection(asset)} onSelectionChange={updateAssetCurrencies} onChange={updateAssetValue} />
              </fieldset>
            ))}
          </div>
          {duplicateAssetIDs(form.assets).size > 0 && <p className="field-error" id="duplicate-asset-names" role="alert">Asset names must be unique, ignoring capitalisation and surrounding spaces.</p>}
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
              <label>User One<input inputMode="decimal" aria-label="User One income" required value={form.income.userOneGBP} onChange={(event) => setForm({ ...form, income: { ...form.income, userOneGBP: event.target.value } })} /></label>
              <label>User Two<input inputMode="decimal" aria-label="User Two income" required value={form.income.userTwoGBP} onChange={(event) => setForm({ ...form, income: { ...form.income, userTwoGBP: event.target.value } })} /></label>
            </div>
          </section>
        </div>
        <section className="panel form-panel">
          <div className="section-heading"><div><p className="eyebrow">Monthly guardrails</p><h2>Spending limits</h2></div><button className="secondary-button" onClick={() => setForm({ ...form, spendingLimits: [...form.spendingLimits, { key: "New limit", amount: "0", currency: "GBP", localID: newLimitID() }] })} type="button">+ Add limit</button></div>
          {form.spendingLimits.length === 0 && <p className="empty-panel">No limits yet. Add one to keep a little space around your spending.</p>}
          <div className="editor-list">
            {form.spendingLimits.map((limit, index) => <div className="limit-editor" key={limit.localID}>
              <label>Limit name<input aria-label={`Spending limit ${index + 1} name`} required value={limit.key} onChange={(event) => updateLimit(limit.localID, (current) => ({ ...current, key: event.target.value }))} /></label>
              <label>Amount<input inputMode="decimal" aria-label={`Spending limit ${index + 1} amount`} required value={limit.amount} onChange={(event) => updateLimit(limit.localID, (current) => ({ ...current, amount: event.target.value }))} /></label>
              <label>Currency<select aria-label={`Spending limit ${index + 1} currency`} value={limit.currency} onChange={(event) => updateLimit(limit.localID, (current) => ({ ...current, currency: event.target.value as "GBP" | "INR" }))}><option>GBP</option><option>INR</option></select></label>
              <button className="icon-button danger" aria-label={`Remove ${limit.key} spending limit`} title={`Remove ${limit.key} spending limit`} onClick={() => setForm({ ...form, spendingLimits: form.spendingLimits.filter((current) => current.localID !== limit.localID) })} type="button"><TrashIcon /></button>
            </div>)}
          </div>
        </section>
        <div className="form-actions"><a className="secondary-button" href="/">Cancel</a><button className="primary-button" disabled={status === "saving"} type="submit">{status === "saving" ? "Saving…" : "Save snapshot"}</button></div>
      </form>
      <footer>Finny · private by design</footer>
    </main>
  );
}

function AssetCurrencyFields({
  asset,
  selection,
  onSelectionChange,
  onChange,
}: {
  asset: Asset;
  selection: AssetCurrencySelection;
  onSelectionChange: (assetID: number, selection: AssetCurrencySelection) => void;
  onChange: (assetID: number, type: "UKGBP" | "INDIAINR", value: string) => void;
}) {
  const valueTypes: AssetValueType[] = selection === "GBP" ? ["UKGBP"] : selection === "INR" ? ["INDIAINR"] : ["UKGBP", "INDIAINR"];
  return (
    <div className="asset-currency-fields">
      <label>Currency<select aria-label={`${asset.name} currency`} value={selection} onChange={(event) => onSelectionChange(asset.id, event.target.value as AssetCurrencySelection)}>
        <option value="GBP">GBP only</option>
        <option value="INR">INR only</option>
        <option value="BOTH">GBP + INR</option>
      </select></label>
      <div className="value-grid">
        {valueTypes.map((type) => {
          const value = asset.values.find((item) => item.type === type);
          const label = type === "UKGBP" ? "United Kingdom · GBP" : "India · INR";
          return <label key={type}>{label}<input inputMode="decimal" aria-label={`${asset.name} ${label} value`} required value={value?.value ?? ""} onChange={(event) => onChange(asset.id, type, event.target.value)} /></label>;
        })}
      </div>
    </div>
  );
}

function TrashIcon() {
  return <svg aria-hidden="true" viewBox="0 0 24 24" focusable="false"><path d="M4 7h16M10 11v6m4-6v6M6 7l1 13h10l1-13M9 7V4h6v3" /></svg>;
}
