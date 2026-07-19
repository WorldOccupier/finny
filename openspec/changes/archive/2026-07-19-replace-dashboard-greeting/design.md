## Context

The dashboard hero currently uses time-specific greeting copy even though the page is a persistent financial overview. The existing hero already has a supporting “Your financial picture” eyebrow, so the change should promote that idea into the main heading and avoid duplicate wording.

## Goals / Non-Goals

**Goals:**

- Display `Your financial picture.` as the dashboard hero heading.
- Use `Dashboard` as the eyebrow label.
- Keep the supporting description and dashboard layout unchanged.
- Update unit and browser assertions to protect the new copy.

**Non-Goals:**

- No dynamic time-of-day or personalization logic.
- No API, routing, styling-system, or data changes.

## Decisions

- Keep the copy as static JSX text because it is a stable product label, not user data.
- Replace the existing eyebrow text rather than leaving duplicate “Your financial picture” text above the heading.
- Update all tests that assert the old greeting so the intended copy is enforced consistently in component and browser coverage.

## Risks / Trade-offs

- [Risk] A stale test or browser assertion could still require the old greeting → Search all source and browser tests for `Good morning` and update every relevant assertion.
- [Risk] A punctuation mismatch could create inconsistent UI expectations → Treat the exact heading string `Your financial picture.` as the canonical copy.

## Migration Plan

No migration is required. This is a frontend copy-only change and can be rolled back by reverting the affected JSX and test assertions.

## Open Questions

- None.
