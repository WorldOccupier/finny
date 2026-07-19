## 1. Inspect current implementation

- [x] 1.1 Review the existing compact-asset-editor implementation and identify the markup causing the name field to occupy a full row.
- [x] 1.2 Confirm the follow-up remains frontend-only and preserves the existing asset API and form state behavior.

## 2. Implement true compact layout

- [x] 2.1 Refactor the asset editor markup into an explicit compact row containing name, currency/value controls, and delete action.
- [x] 2.2 Update wide-screen CSS grid sizing so the asset name no longer spans the full card width.
- [x] 2.3 Update mobile CSS to stack controls cleanly without horizontal scrolling or clipped labels.
- [x] 2.4 Preserve selected-currency rendering, retained values, duplicate-name validation, keyboard navigation, and accessible labels.

## 3. Verify

- [x] 3.1 Update frontend tests for the compact row structure, single/dual currency inputs, removal, validation, and accessibility.
- [x] 3.2 Run `npm test -- --run` from `web/`.
- [x] 3.3 Run `npm run build` from `web/`.
- [x] 3.4 Run formatting/diff checks and visually review desktop and mobile density against the supplied screenshot.
