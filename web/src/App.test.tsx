import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

const empty = { revision: 0, assets: [], currentFxRate: "0", currentTotals: { country: [], combined: [] }, history: [], spendingLimits: [], income: { userOneGBP: "0", userTwoGBP: "0" } };

afterEach(() => {
  cleanup();
  window.history.pushState({}, "", "/");
  vi.restoreAllMocks();
});

describe("browser routes", () => {
  it("renders the dashboard route", async () => {
    window.history.pushState({}, "", "/");
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(empty), { status: 200 })));
    render(<App />);
    await waitFor(() => expect(screen.getByText("Your financial picture starts here")).toBeInTheDocument());
  });

  it("renders the edit route", async () => {
    window.history.pushState({}, "", "/edit");
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(empty), { status: 200 })));
    render(<App />);
    await waitFor(() => expect(screen.getByText("Edit dashboard.")).toBeInTheDocument());
    expect(screen.getByRole("button", { name: "Save snapshot" })).toBeInTheDocument();
  });
});

