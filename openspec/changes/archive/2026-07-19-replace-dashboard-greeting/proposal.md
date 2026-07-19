## Why

The dashboard currently greets the user with “Good morning.”, which is time-specific and less descriptive of the page’s purpose. Replacing it with “Your financial picture.” makes the hero heading clear and useful at any time of day.

## What Changes

- Replace the dashboard hero heading “Good morning.” with “Your financial picture.”.
- Change the eyebrow text to “Dashboard” so the phrase is not duplicated above and below the heading.
- Update unit and browser tests to assert the new heading and remove the old greeting expectation.

## Capabilities

### New Capabilities

- `dashboard-hero-copy`: Stable, time-independent copy for the dashboard hero heading.

### Modified Capabilities

<!-- No existing capability is modified; this is a presentation-only addition. -->

## Impact

- `web/src/features/dashboard/DashboardPage.tsx`
- `web/src/features/dashboard/DashboardPage.test.tsx`
- `web/browser-tests/dashboard.spec.ts`
- No API, persistence, or data-model changes.
