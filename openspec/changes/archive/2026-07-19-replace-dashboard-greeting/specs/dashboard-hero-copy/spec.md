## ADDED Requirements

### Requirement: Stable dashboard hero heading

The dashboard SHALL display `Your financial picture.` as its primary hero heading and SHALL use `Dashboard` as the eyebrow label above it.

#### Scenario: Dashboard with populated data

- **WHEN** the dashboard loads successfully with financial data
- **THEN** the page shows the `Dashboard` eyebrow and the `Your financial picture.` heading

#### Scenario: Empty dashboard

- **WHEN** the dashboard has no snapshots or assets
- **THEN** the page still uses `Your financial picture starts here` for the empty-state heading and does not display `Good morning.`

#### Scenario: Browser rendering

- **WHEN** a browser opens the dashboard route
- **THEN** the visible primary heading is the stable financial-picture copy rather than a time-specific greeting
