import { useState } from "react";
import {
  confirmImport,
  ImportPreview,
  previewImport,
} from "../../api/imports";

export function SpendingPage() {
  const [file, setFile] = useState<File>();
  const [accountId, setAccountId] = useState("checking");
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
        <input
          value={accountId}
          onChange={(event) => setAccountId(event.target.value)}
        />
      </label>
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
              {preview.transactions.map((item) => (
                <tr key={item.id}>
                  <td>{item.sourceRow}</td>
                  <td>{item.date}</td>
                  <td>{item.description}</td>
                  <td>
                    {item.amount} {item.currency}
                  </td>
                </tr>
              ))}
              {preview.invalidRows.map((item) => (
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
    </main>
  );
}
