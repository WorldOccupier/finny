## 1. Inspect and prepare

- [x] 1.1 Review the current Edit Dashboard asset markup, styles, tests, and responsive breakpoints.
- [x] 1.2 Confirm the implementation remains frontend-only and preserves the existing asset API shape.

## 2. Implement compact asset editor

- [x] 2.1 Update the asset editor markup/classes to group name, currency, active values, and remove action compactly.
- [x] 2.2 Update desktop CSS to use a compact responsive grid with usable minimum widths and aligned delete control.
- [x] 2.3 Update mobile CSS so asset controls wrap without horizontal scrolling or clipped content.
- [x] 2.4 Preserve active-currency rendering, retained values, duplicate-name validation, and accessible labels.

## 3. Verify

- [x] 3.1 Add or update frontend tests for single-currency, dual-currency, removal, validation, and accessible controls.
- [x] 3.2 Run `npm test -- --run` from `web/`.
- [x] 3.3 Run `npm run build` from `web/`.
- [x] 3.4 Run formatting and diff checks; manually review desktop and mobile density.
