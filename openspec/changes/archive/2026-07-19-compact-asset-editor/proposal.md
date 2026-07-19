## Why

The Edit Dashboard asset editor uses a tall, heavily padded layout that makes a simple asset update consume too much screen space. A compact, responsive layout will make common edits faster while keeping the currency choice, values, and removal action easy to understand.

## What Changes

- Replace the vertically stacked asset controls with a compact responsive row layout.
- Keep the asset name, currency selector, selected currency values, and remove action visible together on wider screens.
- Collapse the same controls into a compact two-line layout on narrow screens.
- Reduce redundant visual framing, padding, and vertical gaps without removing labels or accessibility metadata.
- Preserve GBP-only, INR-only, and GBP + INR selection behavior.

## Capabilities

### New Capabilities

- `compact-asset-editor`: Compact, responsive asset editing controls for the dashboard editor.

### Modified Capabilities

<!-- No existing OpenSpec capabilities are present in this repository. -->

## Impact

- Frontend Edit Dashboard asset markup and styling in `web/src/features/dashboard/` and `web/src/styles.css`.
- Frontend component and accessibility tests for asset editing.
- No API, persistence, or financial calculation changes.
