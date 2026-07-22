import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SpendingPage } from "./SpendingPage";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("SpendingPage", () => {
  it("renders account, mapping, filters, and empty states", async () => {
    vi.stubGlobal("fetch", vi.fn((url: string) => {
      if (url.includes("/api/accounts")) return Promise.resolve(new Response(JSON.stringify({ accounts: [{ id: "checking", accountLabel: "Checking", owner: "joint" }] })));
      if (url.includes("/api/spending/summary")) return Promise.resolve(new Response(JSON.stringify({ period: "month", summary: [] })));
      return Promise.resolve(new Response(JSON.stringify({ transactions: [], page: 1, pageSize: 25, total: 0 })));
    }));
    render(<SpendingPage />);
    expect(screen.getByRole("heading", { name: "Spending" })).toBeInTheDocument();
    expect(screen.getByText("Column indexes")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("No transactions found.")).toBeInTheDocument());
    expect(screen.getByLabelText("Currency")).toBeInTheDocument();
    expect(screen.getByLabelText("Direction")).toBeInTheDocument();
  });

  it("handles null collection responses without crashing", async () => {
    vi.stubGlobal("fetch", vi.fn((url: string) => {
      if (url.includes("/api/accounts")) return Promise.resolve(new Response(JSON.stringify({ accounts: null })));
      if (url.includes("/api/spending/summary")) return Promise.resolve(new Response(JSON.stringify({ period: "month", summary: null })));
      return Promise.resolve(new Response(JSON.stringify({ transactions: null, page: 1, pageSize: 25, total: 0 })));
    }));
    render(<SpendingPage />);
    await waitFor(() => expect(screen.getByText("No transactions found.")).toBeInTheDocument());
  });
});
