# Finny

Finny is a personal household finance application for tracking net worth, monthly spending limits, and income across the UK and India.

The application is planned as a single repository containing two independently run applications:

```text
finny/
├── server/    # Go HTTP API
└── web/       # React + Vite frontend
```

The web application is served by Vite, while the Go application serves JSON APIs only. During development, both applications run in separate Docker containers, and Vite proxies `/api/*` requests to the Go server.

## Planned features

- Household net-worth snapshots across UK and India.
- Shared asset rows displayed in both country sections.
- GBP and INR values with manually entered exchange rates.
- Historical net-worth graphs with snapshot values preserved over time.
- Configurable monthly spending limits.
- Separate GBP income totals for two household users.

Bank transaction imports, automatic exchange rates, authentication, and remote deployment will be added or decided separately.

## Documentation

- [Web application](web/README.md) — React, Vite, frontend structure, and development workflow.
- [Go server](server/README.md) — HTTP API, persistence, business rules, and server workflow.

## Development model

The frontend and backend run as separate Docker containers:

```text
React/Vite: http://localhost:5173
Go API:     http://localhost:8080
```

Docker Compose will orchestrate the development containers. SQLite data will live in a host-mounted volume so restarting or replacing the server container does not remove the database. The exact commands and configuration will be documented as implementation is added. The Go server will not serve the React build output.
