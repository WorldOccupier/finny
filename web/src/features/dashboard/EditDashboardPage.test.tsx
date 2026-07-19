import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { EditDashboardPage } from "./EditDashboardPage";

const empty = {
  revision: 0,
  assets: [],
  currentFxRate: "0",
  currentTotals: { country: [], combined: [] },
  history: [],
  spendingLimits: [],
  income: { userOneGBP: "0", userTwoGBP: "0" },
};

const populated = {
  revision: 2,
  assets: [{ id: 0, name: "Savings", values: [{ type: "UKGBP", value: "100" }, { type: "INDIAINR", value: "5000" }] }],
  currentFxRate: "100",
  currentTotals: { country: [{ type: "UKGBP", value: "100" }, { type: "INDIAINR", value: "5000" }], combined: [{ currency: "GBP", value: "150" }, { currency: "INR", value: "15000" }] },
  history: [{ id: 1, committedAt: "2026-07-15T13:00:00+01:00", fxRate: "100", assets: [], totals: { country: [{ type: "UKGBP", value: "100" }, { type: "INDIAINR", value: "5000" }], combined: [{ currency: "GBP", value: "150" }, { currency: "INR", value: "15000" }] } }],
  spendingLimits: [],
  income: { userOneGBP: "3000", userTwoGBP: "2500" },
};

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("EditDashboardPage", () => {
  it("loads the complete graph and submits edited assets, FX, income, and limits", async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(populated), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ...populated, revision: 3 }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByDisplayValue("Savings")).toBeInTheDocument());

    fireEvent.change(screen.getByDisplayValue("Savings"), { target: { value: "House savings" } });
    fireEvent.change(screen.getByLabelText("House savings United Kingdom · GBP value"), { target: { value: "200" } });
    fireEvent.change(screen.getByLabelText("Indian rupees per pound"), { target: { value: "105" } });
    fireEvent.change(screen.getByLabelText("User One income"), { target: { value: "3200" } });
    fireEvent.click(screen.getByRole("button", { name: "+ Add limit" }));
    fireEvent.change(screen.getByLabelText("Spending limit 1 name"), { target: { value: "Fun" } });
    fireEvent.change(screen.getByLabelText("Spending limit 1 amount"), { target: { value: "250" } });
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));

    await waitFor(() => expect(screen.getByText("Your snapshot is saved.")).toBeInTheDocument());
    const [, post] = fetchMock.mock.calls;
    expect(post[0]).toBe("/api/dashboard");
    expect(post[1].method).toBe("POST");
    expect(post[1].headers["Idempotency-Key"]).toBeTruthy();
    expect(JSON.parse(post[1].body)).toMatchObject({ revision: 2, fxRate: "105", income: { userOneGBP: "3200" } });
  });

  it("keeps form data after a validation error", async () => {
    vi.stubGlobal("fetch", vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(empty), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: { code: "validation_error", message: "invalid value" } }), { status: 400 })));
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "+ Add asset" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "+ Add asset" }));
    fireEvent.change(screen.getByDisplayValue("New asset"), { target: { value: "Emergency fund" } });
    fireEvent.change(screen.getByLabelText("Emergency fund United Kingdom · GBP value"), { target: { value: "100" } });
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    await waitFor(() => expect(screen.getByText("invalid value")).toBeInTheDocument());
    expect(screen.getByDisplayValue("Emergency fund")).toBeInTheDocument();
  });

  it("offers a reload flow after a revision conflict", async () => {
    vi.stubGlobal("fetch", vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(populated), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: { code: "revision_conflict", message: "stale" } }), { status: 409 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ...populated, revision: 4 }), { status: 200 })));
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByDisplayValue("Savings")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    await waitFor(() => expect(screen.getByRole("button", { name: "Reload latest dashboard" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Reload latest dashboard" }));
    await waitFor(() => expect(screen.getByText("Revision 4")).toBeInTheDocument());
  });

  it("reuses the same idempotency key after a lost response", async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(empty), { status: 200 }))
      .mockRejectedValueOnce(new TypeError("network reset"))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ...populated, revision: 1 }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "+ Add asset" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "+ Add asset" }));
    fireEvent.change(screen.getByDisplayValue("New asset"), { target: { value: "Retry fund" } });
    fireEvent.change(screen.getByLabelText("Retry fund United Kingdom · GBP value"), { target: { value: "100" } });
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    await waitFor(() => expect(screen.getByText("We could not save your dashboard right now.")).toBeInTheDocument());
    const firstKey = fetchMock.mock.calls[1][1].headers["Idempotency-Key"];

    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    await waitFor(() => expect(screen.getByText("Your snapshot is saved.")).toBeInTheDocument());
    expect(fetchMock.mock.calls[2][1].headers["Idempotency-Key"]).toBe(firstKey);
  });

  it("starts a new idempotency operation after the form changes", async () => {
    const saved = { ...populated, revision: 3 };
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(populated), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(saved), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ...saved, revision: 4 }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByDisplayValue("Savings")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    await waitFor(() => expect(screen.getByText("Your snapshot is saved.")).toBeInTheDocument());
    const firstKey = fetchMock.mock.calls[1][1].headers["Idempotency-Key"];
    fireEvent.change(screen.getByDisplayValue("Savings"), { target: { value: "Updated savings" } });
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    await waitFor(() => expect(screen.getByText("Your snapshot is saved.")).toBeInTheDocument());
    expect(fetchMock.mock.calls[2][1].headers["Idempotency-Key"]).not.toBe(firstKey);
  });

  it("keeps a spending-limit input mounted while its name changes", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(empty), { status: 200 })));
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "+ Add limit" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "+ Add limit" }));
    const input = screen.getByLabelText("Spending limit 1 name");
    input.focus();
    fireEvent.change(input, { target: { value: "F" } });
    expect(screen.getByLabelText("Spending limit 1 name")).toBe(input);
    expect(document.activeElement).toBe(input);
  });

  it("supports GBP-only, INR-only, and both-currency assets with a dropdown", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(empty), { status: 200 })));
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "+ Add asset" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "+ Add asset" }));
    const currency = screen.getByRole("combobox", { name: "New asset currency" });
    expect(currency).toHaveValue("GBP");
    expect(screen.queryByLabelText("New asset India · INR value")).not.toBeInTheDocument();
    fireEvent.change(currency, { target: { value: "BOTH" } });
    expect(screen.getByLabelText("New asset India · INR value")).toHaveValue("");
    fireEvent.change(currency, { target: { value: "INR" } });
    expect(screen.queryByLabelText("New asset United Kingdom · GBP value")).not.toBeInTheDocument();
    expect(screen.getByLabelText("New asset India · INR value")).toBeInTheDocument();
  });

  it("converts an existing asset between currency memberships", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(populated), { status: 200 })));
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByDisplayValue("Savings")).toBeInTheDocument());
    const currency = screen.getByRole("combobox", { name: "Savings currency" });
    fireEvent.change(currency, { target: { value: "GBP" } });
    expect(screen.queryByLabelText("Savings India · INR value")).not.toBeInTheDocument();
    fireEvent.change(currency, { target: { value: "INR" } });
    expect(screen.queryByLabelText("Savings United Kingdom · GBP value")).not.toBeInTheDocument();
    expect(screen.getByLabelText("Savings India · INR value")).toBeInTheDocument();
  });

  it("rejects duplicate asset names without submitting", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(empty), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "+ Add asset" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "+ Add asset" }));
    fireEvent.click(screen.getByRole("button", { name: "+ Add asset" }));
    const names = screen.getAllByRole("textbox", { name: /asset \d+ name/i });
    fireEvent.change(names[0], { target: { value: "Savings" } });
    fireEvent.change(names[1], { target: { value: "  savings " } });
    fireEvent.change(screen.getByLabelText("Savings United Kingdom · GBP value"), { target: { value: "100" } });
    fireEvent.change(screen.getAllByLabelText(/United Kingdom · GBP value/)[1], { target: { value: "200" } });
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    expect(screen.getByRole("alert")).toHaveTextContent("Asset names must be unique");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(names[1]).toHaveValue("  savings ");
  });

  it("allows distinct asset names", async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(empty), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ...populated, revision: 1 }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "+ Add asset" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "+ Add asset" }));
    fireEvent.click(screen.getByRole("button", { name: "+ Add asset" }));
    const names = screen.getAllByRole("textbox", { name: /asset \d+ name/i });
    fireEvent.change(names[0], { target: { value: "Savings" } });
    fireEvent.change(names[1], { target: { value: "Brokerage" } });
    fireEvent.change(screen.getByLabelText("Savings United Kingdom · GBP value"), { target: { value: "100" } });
    fireEvent.change(screen.getByLabelText("Brokerage United Kingdom · GBP value"), { target: { value: "200" } });
    fireEvent.click(screen.getByRole("button", { name: "Save snapshot" }));
    await waitFor(() => expect(screen.getByText("Your snapshot is saved.")).toBeInTheDocument());
  });

  it("uses accessible trash buttons for assets and limits", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(populated), { status: 200 })));
    render(<EditDashboardPage />);
    await waitFor(() => expect(screen.getByDisplayValue("Savings")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "+ Add limit" }));
    expect(screen.getByRole("button", { name: "Remove Savings" })).toHaveAttribute("title", "Remove Savings");
    expect(screen.getByRole("button", { name: "Remove New limit spending limit" })).toBeInTheDocument();
  });
});
