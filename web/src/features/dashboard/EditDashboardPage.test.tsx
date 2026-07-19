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
    fireEvent.change(screen.getByLabelText("House savings UK GBP value"), { target: { value: "200" } });
    fireEvent.change(screen.getByLabelText("Indian rupees per pound"), { target: { value: "105" } });
    fireEvent.change(screen.getByLabelText("User one income"), { target: { value: "3200" } });
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
});
