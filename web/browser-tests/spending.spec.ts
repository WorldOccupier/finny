import { expect, test } from "@playwright/test";

test.describe("spending browser workflow", () => {
  test("previews an invalid CSV and exposes all summary periods", async ({ page }) => {
    await page.goto("/spending");
    await expect(page.getByRole("heading", { name: "Spending" })).toBeVisible();
    await page.getByLabel("Statement file").setInputFiles({
      name: "statement.csv",
      mimeType: "text/csv",
      buffer: Buffer.from(
        "date,description,amount,currency\n2026-07-20,Coffee,-3.50,GBP\ninvalid,Missing date,2,GBP\n",
      ),
    });
    await page.getByRole("button", { name: "Preview import" }).click();
    await expect(page.getByRole("heading", { name: "Preview" })).toBeVisible();
    await expect(page.getByText(/1 valid, 1 invalid/)).toBeVisible();
    await expect(page.getByText("invalid date", { exact: false })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Summary" })).toBeVisible();
    await page.getByLabel("Period").selectOption("year");
    await expect(page.getByLabel("Period")).toHaveValue("year");
  });

  test("searches the transaction table", async ({ page }) => {
    await page.goto("/spending");
    await page.getByLabel("Search").fill("coffee");
    await expect(page.getByText("No transactions found.")).toBeVisible();
  });
});
