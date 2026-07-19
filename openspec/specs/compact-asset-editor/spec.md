# Compact Asset Editor

## Purpose

Define the compact, responsive asset-editing experience in the Edit Dashboard while preserving the existing asset data and accessibility behavior.

## Requirements

### Requirement: Compact asset controls

The Edit Dashboard SHALL present each asset's name, selected currency option, active currency value inputs, and remove action in a genuinely dense grouped layout. On wide screens, the primary controls SHALL occupy one compact row rather than placing the asset name in a full-width row above the currency and value controls.

#### Scenario: Single-currency asset on a wide screen

- **WHEN** an asset is configured as GBP-only or INR-only
- **THEN** the editor shows the asset name, one currency selector, one corresponding value input, and the remove action in one compact horizontal group without rendering an unused second currency field

#### Scenario: Dual-currency asset on a wide screen

- **WHEN** an asset is configured for both GBP and INR
- **THEN** the editor keeps the name, currency selector, both value inputs, and remove action in one compact grouped row with usable minimum widths

#### Scenario: Asset editor on a narrow screen

- **WHEN** the available viewport is narrow
- **THEN** the controls wrap into a compact readable layout without horizontal scrolling or clipped labels

### Requirement: Preserve asset editing behavior and accessibility

The compact layout SHALL preserve existing asset editing behavior, including currency selection, value editing, removal, duplicate-name validation, keyboard navigation, and accessible labels.

#### Scenario: Change asset currency selection

- **WHEN** the user changes an asset from GBP-only to INR-only, or to both currencies
- **THEN** the corresponding value inputs appear or disappear and existing values for retained currencies remain available

#### Scenario: Remove an asset

- **WHEN** the user activates the asset remove button
- **THEN** only that asset is removed from the editable form and the remove button remains identifiable by its accessible name and trash icon

#### Scenario: Duplicate asset names

- **WHEN** two asset names match after trimming surrounding spaces and ignoring capitalisation
- **THEN** the existing validation message and invalid state remain visible and submission is prevented

### Requirement: Visual density regression coverage

The frontend SHALL include automated coverage for the compact asset editor's single-row desktop structure and responsive class structure without changing the dashboard API contract.

#### Scenario: Existing asset form tests

- **WHEN** the frontend test suite renders the Edit Dashboard
- **THEN** it verifies the compact asset row contains the name, currency dropdown, active value inputs, and accessible remove control

#### Scenario: Production styling validation

- **WHEN** the frontend formatter, tests, and production build run
- **THEN** the compact asset editor styles compile successfully with no API or type errors
