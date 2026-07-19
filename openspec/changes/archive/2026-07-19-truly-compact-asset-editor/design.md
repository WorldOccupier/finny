## Context

The first compact asset-editor change reduced padding and placed currency/value controls together, but the screenshot shows the name input still occupies a complete row and leaves a large amount of unused space. The follow-up should make the wide-screen asset editor a true dense row while keeping mobile usability and the existing controlled form behavior.

## Goals / Non-Goals

**Goals:**

- Fit the asset name, currency selector, active values, and delete action into one compact desktop row.
- Give each control a proportionate width rather than allowing the name to consume the whole card.
- Keep the layout readable and touch-friendly on mobile through a responsive fallback.
- Preserve existing semantics, accessible labels, validation, and currency-specific value rendering.

**Non-Goals:**

- No API, domain, persistence, or calculation changes.
- No removal of visible field labels.
- No accordion or hidden asset fields.

## Decisions

- Change the asset editor markup to have a single `.asset-editor-row` wrapper containing the name field, currency/value field group, and delete button. This makes the intended compact layout explicit instead of relying on unrelated nested flex rules.
- Use CSS grid columns such as `minmax(150px, 1fr) minmax(150px, .8fr) minmax(180px, 1.2fr) auto` on wide screens. The value group will use an inner grid for one or two active values.
- Remove the fieldset legend from the visible layout and use a visually subtle asset index only if needed for orientation; the accessible labels remain on each input.
- At the mobile breakpoint, switch the row to one column and keep the delete action aligned with the asset name row so the form remains compact without horizontal scrolling.
- Prefer CSS-only responsive behavior. A collapsible design was considered but rejected because it hides editable controls and adds interaction cost.

## Risks / Trade-offs

- [Risk] Very long names can compress neighboring controls → Apply `minmax(0, 1fr)`, input overflow handling, and mobile stacking.
- [Risk] Two active currency values need more width → Let the value group span available space and use compact minimum widths; stack only at the narrow breakpoint.
- [Risk] Removing visible legends can reduce orientation → Retain clear field labels and accessible names, and keep the asset group border/background subtle.

## Migration Plan

No data migration is required. Deploy the frontend markup/CSS/test changes together. Rollback is a revert of the frontend-only commit.

## Open Questions

- None. The screenshot establishes that the desktop name field must no longer occupy a full-width row.
