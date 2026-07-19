## Why

The previous compacting pass reduced padding but still leaves each asset editor visually large: the asset name occupies a full row and the currency/value controls begin below it. The screenshot shows that the editor still hogs vertical space, so the controls need to become a genuinely dense, single-row desktop editor with a deliberate mobile fallback.

## What Changes

- Rework the asset editor so the name, currency selector, active values, and delete action share one compact desktop row.
- Remove redundant asset legends and excess nested framing where it does not improve usability.
- Keep inputs short and proportionate instead of stretching the asset name across the entire card.
- Preserve a readable stacked layout for narrow screens.
- Keep existing currency selection, validation, removal, keyboard, and accessibility behavior.

## Capabilities

### New Capabilities

<!-- No new capability; this is a refinement of the existing asset-editor contract. -->

### Modified Capabilities

- `compact-asset-editor`: Require a genuinely dense single-row desktop layout while preserving the existing editing behavior.

## Impact

- Frontend asset editor markup and styles in `web/src/features/dashboard/EditDashboardPage.tsx` and `web/src/styles.css`.
- Frontend tests covering compact structure and responsive behavior.
- No API, persistence, or financial calculation changes.
