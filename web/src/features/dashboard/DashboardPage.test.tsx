import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { DashboardPage } from "./DashboardPage";

const response = {
  revision: 4,
  assets: [{ id: 0, name: "Savings", values: [{ type: "UKGBP", value: "1000" }, { type: "INDIAINR", value: "50000" }] }],
  currentFxRate: "100",
  currentTotals: { country: [{ type: "UKGBP", value: "1000" }, { type: "INDIAINR", value: "50000" }], combined: [{ currency: "GBP", value: "1500" }, { currency: "INR", value: "150000" }] },
  history: [{ id: 1, committedAt: "2026-01-15T12:00:00Z", fxRate: "100", assets: [], totals: { country: [{ type: "UKGBP", value: "1000" }, { type: "INDIAINR", value: "50000" }], combined: [{ currency: "GBP", value: "1500" }, { currency: "INR", value: "150000" }] } }],
  spendingLimits: [{ key: "Fun", amount: "200", currency: "GBP" }],
  income: { userOneGBP: "3000", userTwoGBP: "2500" },
};

function mockDashboardResponse(body: unknown = response) {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(body), { status: 200 })));
}

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("DashboardPage", () => {
  it("renders the complete dashboard response", async () => {
    mockDashboardResponse();
    render(<DashboardPage />);

    expect(screen.getByText("Loading your dashboard")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("Good morning.")).toBeInTheDocument());
    expect(screen.getAllByText("Savings")).toHaveLength(2);
    expect(screen.getByText("Spending limits")).toBeInTheDocument();
    expect(screen.getByText("Monthly income")).toBeInTheDocument();
    expect(screen.getByLabelText("Net worth values in GBP")).toBeInTheDocument();
    expect(screen.getAllByText("£1,500.00").length).toBeGreaterThanOrEqual(2);
  });

  it("renders the empty state", async () => {
    mockDashboardResponse({
      ...response,
      revision: 0,
      assets: [],
      currentTotals: { country: [], combined: [] },
      history: [],
      spendingLimits: [],
      income: { userOneGBP: "0", userTwoGBP: "0" },
      currentFxRate: "0",
    });
    render(<DashboardPage />);
    await waitFor(() => expect(screen.getByText("Your financial picture starts here")).toBeInTheDocument());
  });

  it("renders the error state when the API fails", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("{}", { status: 503 })));
    render(<DashboardPage />);
    await waitFor(() => expect(screen.getByText("Your dashboard is taking a moment")).toBeInTheDocument());
  });

  it("switches the combined display between GBP and INR", async () => {
    mockDashboardResponse();
    render(<DashboardPage />);
    await waitFor(() => expect(screen.getByText("Good morning.")).toBeInTheDocument());
    const inrButton = screen.getAllByRole("button", { name: "INR" })[0];
    fireEvent.click(inrButton);
    expect(screen.getAllByText("₹150,000.00").length).toBeGreaterThanOrEqual(2);
    expect(screen.getByRole("img", { name: "Net worth history in INR" })).toBeInTheDocument();
  });

  it.each([
    ["malformed response", { ...response, assets: [{ id: 0, name: "Savings", values: [{ type: "UNKNOWN", value: "1000" }] }] }],
    ["invalid decimal", { ...response, currentFxRate: "not-a-decimal" }],
    ["missing current totals", { ...response, currentTotals: { country: [], combined: [] } }],
    ["incomplete combined totals", { ...response, currentTotals: { ...response.currentTotals, combined: [{ currency: "GBP", value: "1500" }] } }],
    ["duplicate combined totals", { ...response, currentTotals: { ...response.currentTotals, combined: [{ currency: "GBP", value: "1500" }, { currency: "GBP", value: "1600" }] } }],
    ["incomplete country totals", { ...response, currentTotals: { ...response.currentTotals, country: [{ type: "UKGBP", value: "1000" }] } }],
    ["duplicate country totals", { ...response, currentTotals: { ...response.currentTotals, country: [{ type: "UKGBP", value: "1000" }, { type: "UKGBP", value: "1100" }] } }],
    ["missing snapshot totals", { ...response, history: [{ ...response.history[0], totals: { country: [], combined: [] } }] }],
    ["date-only timestamp", { ...response, history: [{ ...response.history[0], committedAt: "2026-01-15" }] }],
    ["timezone-less timestamp", { ...response, history: [{ ...response.history[0], committedAt: "2026-01-15T12:00:00" }] }],
    ["invalid timestamp shape", { ...response, history: [{ ...response.history[0], committedAt: "15/01/2026 12:00:00+00:00" }] }],
  ])("renders the error state for a %s", async (_description, body) => {
    mockDashboardResponse(body);
    render(<DashboardPage />);

    await waitFor(() => expect(screen.getByText("Your dashboard is taking a moment")).toBeInTheDocument());
    expect(screen.queryByText("Good morning.")).not.toBeInTheDocument();
  });

  it("exposes the selected currency to keyboard and assistive technology users", async () => {
    mockDashboardResponse();
    render(<DashboardPage />);
    await waitFor(() => expect(screen.getByText("Good morning.")).toBeInTheDocument());

    const gbpButton = screen.getAllByRole("button", { name: "GBP" })[0];
    const inrButton = screen.getAllByRole("button", { name: "INR" })[0];
    expect(gbpButton).toHaveAttribute("aria-pressed", "true");
    expect(inrButton).toHaveAttribute("aria-pressed", "false");

    inrButton.focus();
    expect(document.activeElement).toBe(inrButton);
    fireEvent.click(inrButton);
    expect(inrButton).toHaveAttribute("aria-pressed", "true");
    expect(gbpButton).toHaveAttribute("aria-pressed", "false");
  });
});
