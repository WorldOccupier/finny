import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { Snapshot } from "../../api/dashboard";
import { HistoryChart } from "./HistoryChart";

const history: Snapshot[] = [
  {
    id: 1,
    committedAt: "2026-01-15T12:00:00Z",
    fxRate: "100",
    assets: [],
    totals: {
      country: [],
      combined: [
        { currency: "GBP", value: "150" },
        { currency: "INR", value: "15000" },
      ],
    },
  },
  {
    id: 2,
    committedAt: "2026-02-15T12:00:00Z",
    fxRate: "110",
    assets: [],
    totals: {
      country: [],
      combined: [
        { currency: "GBP", value: "175" },
        { currency: "INR", value: "19250" },
      ],
    },
  },
];

describe("HistoryChart", () => {
  it("uses each snapshot's frozen total for the selected currency", () => {
    const { rerender } = render(<HistoryChart history={history} currency="GBP" />);

    expect(screen.getAllByText("£175.00").length).toBeGreaterThan(0);
    expect(screen.queryByText("£19,250.00")).not.toBeInTheDocument();

    rerender(<HistoryChart history={history} currency="INR" />);

    expect(screen.getAllByText("₹19,250.00").length).toBeGreaterThan(0);
    expect(screen.queryByText("₹175.00")).not.toBeInTheDocument();
  });
});
