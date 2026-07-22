import { useEffect, useState } from "react";
import {
  confirmImport,
  ImportPreview,
  listAccounts,
  listTransactions,
  previewImport,
  spendingSummary,
  TransactionPage,
  TransactionSummary,
} from "../../api/imports";

export function SpendingPage() {
  const [file, setFile] = useState<File>();
  const [accountId, setAccountId] = useState("checking");
  const [user, setUser] = useState("user_one");
  const [accounts, setAccounts] = useState<{ id: string; accountLabel: string; owner: string }[]>([]);
  const [preview, setPreview] = useState<ImportPreview>();
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [mapping, setMapping] = useState({
    date: 0,
    description: 1,
    amount: 2,
    debit: -1,
    credit: -1,
    currency: 3,
    reference: 4,
  });
  const [search, setSearch] = useState("");
  const [currency, setCurrency] = useState("");
  const [direction, setDirection] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [page, setPage] = useState(1);
  const [transactions, setTransactions] = useState<TransactionPage>();
  const [period, setPeriod] = useState("month");
  const [summary, setSummary] = useState<TransactionSummary[]>([]);

  useEffect(() => {
    const params = new URLSearchParams({ q: search, user, page: String(page), pageSize: "25" });
    if (accountId) params.set("accountId", accountId);
    if (currency) params.set("currency", currency);
    if (direction) params.set("direction", direction);
    if (from) params.set("from", from);
    if (to) params.set("to", to);
    listTransactions(params).then(setTransactions).catch(() => setTransactions(undefined));
  }, [search, user, accountId, currency, direction, from, to, page]);

  useEffect(() => {
    listAccounts(user).then((result) => setAccounts(result.accounts ?? [])).catch(() => setAccounts([]));
  }, [user]);

  useEffect(() => {
    spendingSummary(period).then((result) => setSummary(result.summary ?? [])).catch(() => setSummary([]));
  }, [period]);

  const upload = async () => {
    if (!file) return;
    setError("");
    setMessage("");
    try {
      setPreview(await previewImport(file, accountId, mapping));
    } catch (value) {
      setError(value instanceof Error ? value.message : "Preview failed");
    }
  };

  const confirm = async () => {
    if (!preview) return;
    try {
      const result = await confirmImport(preview.token);
      setMessage(`Imported ${result.importedRows} rows`);
    } catch (value) {
      setError(value instanceof Error ? value.message : "Import failed");
    }
  };

  const field = (name: keyof typeof mapping) => (
    <label key={name}>
      {name}
      <input
        type="number"
        value={mapping[name]}
        onChange={(event) =>
          setMapping({ ...mapping, [name]: Number(event.target.value) })
        }
      />
    </label>
  );

  return (
    <main className="page">
      <h1>Spending</h1>
      <p>Import a CSV or XLSX statement and review it before saving.</p>
      <label>
        Account
        <select value={accountId} onChange={(event) => setAccountId(event.target.value)}>
          {accounts.map((account) => <option key={account.id} value={account.id}>{account.accountLabel} ({account.owner})</option>)}
          {!accounts.length && <option value="checking">Checking</option>}
        </select>
      </label>
      <label>User <select value={user} onChange={(event) => setUser(event.target.value)}><option value="user_one">User one</option><option value="user_two">User two</option></select></label>
      <label>
        Statement file
        <input
          type="file"
          accept=".csv,.xlsx"
          onChange={(event) => setFile(event.target.files?.[0])}
        />
      </label>
      <fieldset>
        <legend>Column indexes</legend>
        {Object.keys(mapping).map((name) => field(name as keyof typeof mapping))}
      </fieldset>
      <button onClick={upload} disabled={!file}>
        Preview import
      </button>
      {error && <p role="alert">{error}</p>}
      {message && <p role="status">{message}</p>}
      {preview && (
        <section>
          <h2>Preview</h2>
          <p>
            {preview.validRows} valid, {preview.invalidCount} invalid, {" "}
            {preview.duplicateCount ?? 0} duplicate
          </p>
          <table>
            <thead>
              <tr>
                <th>Row</th>
                <th>Date</th>
                <th>Description</th>
                <th>Amount</th>
              </tr>
            </thead>
            <tbody>
              {(preview.transactions ?? []).map((item) => (
                <tr key={item.id}>
                  <td>{item.sourceRow}</td>
                  <td>{item.date}</td>
                  <td>{item.description}</td>
                  <td>
                    {item.amount} {item.currency}
                  </td>
                </tr>
              ))}
              {(preview.invalidRows ?? []).map((item) => (
                <tr key={`invalid-${item.sourceRow}`}>
                  <td>{item.sourceRow}</td>
                  <td colSpan={3}>{item.message}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <button onClick={confirm}>Confirm import</button>
        </section>
      )}
      <section>
        <h2>Summary</h2>
        <label>
          Period
          <select value={period} onChange={(event) => setPeriod(event.target.value)}>
            {['day', 'week', 'month', 'year'].map((value) => <option key={value}>{value}</option>)}
          </select>
        </label>
        {summary.map((item) => <p key={item.currency}>{item.amount} {item.currency}</p>)}
      </section>
      <section>
        <h2>Transactions</h2>
        <label>
          Search
          <input value={search} onChange={(event) => setSearch(event.target.value)} />
        </label>
        <label>From <input type="date" value={from} onChange={(event) => setFrom(event.target.value)} /></label>
        <label>To <input type="date" value={to} onChange={(event) => setTo(event.target.value)} /></label>
        <label>Currency <select value={currency} onChange={(event) => setCurrency(event.target.value)}><option value="">All</option><option value="GBP">GBP</option><option value="INR">INR</option></select></label>
        <label>Direction <select value={direction} onChange={(event) => setDirection(event.target.value)}><option value="">All</option><option value="debit">Debit</option><option value="credit">Credit</option></select></label>
        {transactions?.transactions?.length ? <table><thead><tr><th>Date</th><th>Description</th><th>Amount</th></tr></thead><tbody>{transactions.transactions.map((item) => <tr key={item.id}><td>{item.date}</td><td>{item.description}</td><td>{item.amount} {item.currency}</td></tr>)}</tbody></table> : <p>No transactions found.</p>}
        <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page === 1}>Previous</button>
        <span>Page {page}</span>
        <button onClick={() => setPage(page + 1)} disabled={!transactions || page * transactions.pageSize >= transactions.total}>Next</button>
      </section>
    </main>
  );
}
