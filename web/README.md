# Finny Web

The `web/` directory contains Finny’s browser application, built with React, TypeScript, and Vite.

## Responsibilities

The web application owns:

- Page layout and navigation.
- Net-worth, income, and spending-limit forms.
- Dashboard cards, country sections, and historical charts.
- Client-side loading, validation feedback, and error states.
- Calling the Go server through HTTP/JSON APIs.

It does not own persistent business data, database access, exchange-rate calculations, or authorization rules. Those responsibilities belong to the Go server.

## Planned structure

```text
web/
├── src/
│   ├── api/                    # HTTP client functions
│   ├── App.tsx                 # / and /edit route selection
│   ├── components/             # Shared UI components
│   ├── features/
│   │   ├── dashboard/
│   │   ├── net-worth/
│   │   ├── income/
│   │   └── spending-limits/
│   └── pages/
├── package.json
├── vite.config.ts
└── README.md
```

Frontend code should be organized primarily by feature. Shared visual components belong in `components/`; API calls belong in `api/`; business persistence logic should not be duplicated in React components.

## Development

Vite serves the frontend in its own Docker container, independently from the Go server container:

```text
Web app:  http://localhost:5173
Go API:   http://localhost:8080
```

During development, Docker Compose runs Vite and Go as separate services. The Vite development server proxies requests beginning with `/api/` to the Go server service. React code should therefore use relative paths, for example:

```text
/api/dashboard
/api/net-worth
/api/income
/api/spending-limits
```

This keeps browser requests same-origin from the frontend’s perspective and avoids embedding environment-specific API hosts throughout the UI.

## Planned commands

Once the frontend is implemented, the expected development commands will be:

```bash
docker compose up web server
```

The web container runs Vite with the source tree mounted for hot reload. The production build remains a Vite concern. The Go server will not serve `web/dist`.

The `/edit` route loads the complete dashboard graph into a controlled form. Saving submits the current revision and one persisted `Idempotency-Key` for that form operation; retrying after a lost response reuses the key, while a successful response or intentional form change starts a new operation. Successful responses replace the form state with the committed dashboard. Validation errors keep the edited values in place, while revision conflicts offer an explicit reload of the latest dashboard.

The `/spending` route uploads CSV/XLSX statements, previews valid and invalid rows, and confirms an import through the statement API.

Browser-level checks use Playwright against a real Go server and Vite process:

```bash
npm run test:browser
```

The clean-install frontend checks pass with `npm ci --ignore-scripts`. A clean isolated copy was also verified with `docker compose up web server --build -d`; the Go health endpoint and Vite HTML shell both responded successfully. Temporary host ports may be needed when the default development ports are already occupied. `docker compose config` is used to validate the Compose configuration.

## UI principles

- Make the combined current net worth prominent.
- Present UK and India in separate tabs or equivalent sections.
- Keep country names in headings, not repeated in asset row labels.
- Show the correct currency for each country section.
- Keep forms clear about whether a value is historical, current, or a monthly limit.
- Keep decimal values as strings in form state and API payloads; do not convert financial values through floating point.
