import { expect, test } from "@playwright/test";

test.describe("dashboard browser workflow", () => {
  test.describe.configure({ mode: "serial" });
  test("navigates to the empty dashboard and editor through the Vite proxy", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByRole("heading", { name: "Your financial picture starts here" })).toBeVisible();
    await page.getByRole("button", { name: "Open editor" }).click();
    await expect(page).toHaveURL(/\/edit$/);
    await expect(page.getByRole("heading", { name: "Edit dashboard." })).toBeVisible();
    await expect(page.getByRole("button", { name: "Save snapshot" })).toBeVisible();
    await expect(page.getByRole("button", { name: "+ Add asset" })).toBeVisible();
    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.getByRole("button", { name: "Save snapshot" })).toBeVisible();
    await expect(page.getByLabel("Indian rupees per pound")).toBeVisible();
  });

  test("creates a first snapshot, carries values forward, and retains removed history", async ({ page }) => {
    await page.goto("/edit");
    await page.getByRole("button", { name: "+ Add asset" }).click();
    await page.getByLabel("Asset name").fill("Emergency fund");
    await page.getByLabel("Emergency fund UK GBP value").fill("1000");
    await page.getByLabel("Emergency fund India INR value").fill("50000");
    await page.getByLabel("Indian rupees per pound").fill("100");
    await page.getByRole("button", { name: "Save snapshot" }).click();
    await expect(page.getByText("Your snapshot is saved.")).toBeVisible();

    let omitIndiaValue = true;
    await page.route("**/api/dashboard", async (route) => {
      if (route.request().method() === "POST" && omitIndiaValue) {
        omitIndiaValue = false;
        const requestBody = JSON.parse(route.request().postData() ?? "{}");
        requestBody.assets[0].values = requestBody.assets[0].values.filter(
          (value: { type: string }) => value.type !== "INDIAINR",
        );
        await route.continue({ postData: JSON.stringify(requestBody) });
        return;
      }
      await route.continue();
    });

    await page.getByLabel("Indian rupees per pound").fill("105");
    await page.getByRole("button", { name: "Save snapshot" }).click();
    await expect(page.getByText("Your snapshot is saved.")).toBeVisible();
    await expect(page.getByLabel("Emergency fund India INR value")).toHaveValue("50000");

    await page.getByRole("button", { name: "Remove" }).click();
    await page.getByLabel("Indian rupees per pound").fill("110");
    await page.getByRole("button", { name: "Save snapshot" }).click();
    await expect(page.getByText("Your snapshot is saved.")).toBeVisible();

    await page.goto("/");
    await expect(page.getByRole("heading", { name: "Good morning." })).toBeVisible();
    await expect(page.getByText(/No .* assets yet/)).toHaveCount(2);
    await expect(page.getByRole("img", { name: /Net worth history/ })).toBeVisible();
    await expect(page.getByText("3 snapshots")).toBeVisible();
  });

  test("replays a committed save when the first response is lost", async ({ page }) => {
    let firstPost = true;
    await page.route("**/api/dashboard", async (route) => {
      if (route.request().method() === "POST" && firstPost) {
        firstPost = false;
        const response = await route.fetch();
        await response.body();
        await route.abort();
        return;
      }
      await route.continue();
    });

    await page.goto("/edit");
    await page.getByRole("button", { name: "+ Add asset" }).click();
    await page.getByLabel("Asset name").fill("Retry fund");
    await page.getByLabel("Retry fund UK GBP value").fill("250");
    await page.getByLabel("Retry fund India INR value").fill("25000");
    await page.getByLabel("Indian rupees per pound").fill("100");
    await page.getByRole("button", { name: "Save snapshot" }).click();
    await expect(page.getByText("We could not save your dashboard right now.")).toBeVisible();

    await page.getByRole("button", { name: "Save snapshot" }).click();
    await expect(page.getByText("Your snapshot is saved.")).toBeVisible();
    await page.goto("/");
    await expect(page.getByText("Retry fund")).toHaveCount(2);
    await expect(page.getByText("4 snapshots")).toBeVisible();
  });
});
