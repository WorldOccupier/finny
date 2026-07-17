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

Vite serves the frontend independently from the Go server:

```text
Web app:  http://localhost:5173
Go API:   http://localhost:8080
```

During development, the Vite development server will proxy requests beginning with `/api/` to the Go server. React code should therefore use relative paths, for example:

```text
/api/dashboard
/api/net-worth
/api/income
/api/spending-limits
```

This keeps browser requests same-origin from the frontend’s perspective and avoids embedding environment-specific API hosts throughout the UI.

## Planned commands

Once the frontend is implemented, the expected commands will be:

```bash
npm install
npm run dev
npm run build
npm run test
```

The production build remains a Vite concern. The Go server will not serve `web/dist`.

## UI principles

- Make the combined current net worth prominent.
- Present UK and India in separate tabs or equivalent sections.
- Keep country names in headings, not repeated in asset row labels.
- Show the correct currency for each country section.
- Keep forms clear about whether a value is historical, current, or a monthly limit.
