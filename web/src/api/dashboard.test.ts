import { afterEach, describe, expect, it, vi } from "vitest";
import { DashboardApiError, saveDashboard } from "./dashboard";

const request = { revision: 0, assets: [], fxRate: "1", spendingLimits: [], income: { userOneGBP: "0", userTwoGBP: "0" } };
const response = { revision: 0, assets: [], currentFxRate: "0", currentTotals: { country: [], combined: [] }, history: [], spendingLimits: [], income: { userOneGBP: "0", userTwoGBP: "0" } };

afterEach(() => vi.restoreAllMocks());

describe("dashboard API client", () => {
  it("posts the complete request with an idempotency key and returns the response", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(response), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    await expect(saveDashboard(request, "retry-key")).resolves.toEqual(response);
    expect(fetchMock).toHaveBeenCalledWith("/api/dashboard", expect.objectContaining({ method: "POST", body: JSON.stringify(request), headers: expect.objectContaining({ "Idempotency-Key": "retry-key" }) }));
  });

  it("exposes conflict status and server error details", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code: "revision_conflict", message: "Reload required" } }), { status: 409 })));
    await expect(saveDashboard(request, "retry-key")).rejects.toMatchObject({ status: 409, code: "revision_conflict", message: "Reload required" });
  });

  it("rejects malformed successful responses", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("{}", { status: 200 })));
    await expect(saveDashboard(request, "retry-key")).rejects.toThrow("saved dashboard response was not valid");
  });
});
