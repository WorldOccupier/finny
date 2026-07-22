export interface ImportedTransaction { id: string; accountId: string; date: string; amount: string; currency: "GBP" | "INR"; description: string; reference?: string; sourceRow: number }
export interface InvalidImportRow { sourceRow: number; message: string }
export interface ImportPreview { token: string; checksum: string; periodStart: string; periodEnd: string; transactions: ImportedTransaction[]; invalidRows: InvalidImportRow[]; validRows: number; invalidCount: number; duplicateCount?: number }

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body?.error?.message ?? "Request failed");
  return body as T;
}

export function previewImport(file: File, accountId: string, mapping: Record<string, number>): Promise<ImportPreview> {
  const form = new FormData(); form.append("file", file); form.append("accountId", accountId); form.append("importedBy", "user_one");
  Object.entries(mapping).forEach(([key, value]) => form.append(key, String(value)));
  return request<ImportPreview>("/api/statements/preview", { method: "POST", body: form });
}

export function confirmImport(token: string): Promise<{ statement: unknown; importedRows: number }> {
  return request("/api/statements/confirm", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ token }) });
}
