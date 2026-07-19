## Context

The Edit Dashboard currently renders each asset as a padded fieldset with the name, currency selector, and value inputs stacked vertically. This is readable but creates a tall page, especially when several assets are present. The change is frontend-only and must preserve the existing asset data model, currency-selection behavior, validation, and accessible controls.

## Goals / Non-Goals

**Goals:**

- Present the asset name, currency selection, active value inputs, and remove action in a compact arrangement.
- Keep all active controls visible on desktop and make the layout wrap cleanly on mobile.
- Reduce redundant padding, gaps, and framing while maintaining clear grouping and focus states.
- Preserve semantic labels, keyboard access, screen-reader names, and existing form behavior.

**Non-Goals:**

- No API or persistence changes.
- No changes to GBP/INR calculations or asset validation rules.
- No collapsible asset editor state in this change.

## Decisions

- Use a responsive CSS grid for each asset editor. The first row will prioritize the name, currency selector, and delete control; active currency values will occupy remaining columns.
- Keep the currency selector as the source of truth for which value inputs render. This avoids showing empty zero-like fields for currencies not selected by the user.
- Retain a compact bordered container rather than removing grouping entirely. This preserves scanability between multiple assets while reducing visual weight.
- Use CSS media queries to stack controls on narrow screens rather than adding JavaScript viewport logic.
- Keep the existing trash icon button in the top-right, with a grid-area/alignment rule that prevents it from floating above the field row.

## Risks / Trade-offs

- [Risk] Long asset names may squeeze the currency and value controls → Use minimum widths and allow the name field to shrink/wrap within the grid.
- [Risk] A dense desktop row may feel cramped with both currencies selected → Give each active value a usable minimum width and allow the value area to span the available row.
- [Risk] Reduced labels could weaken discoverability → Keep visible field labels and existing accessible names; only remove redundant asset numbering/framing.

## Migration Plan

No data migration or rollout step is required. Deploy the frontend changes together; rollback is a frontend asset-editor CSS/markup revert.

## Open Questions

- None for implementation. The compact responsive row is the selected direction from exploration.
