# Finny Agent Instructions

## Source of truth

Before making implementation changes, read:

1. `docs/implementation-plan.md`
2. `docs/decisions.md`
3. The relevant `README.md` files

The implementation plan is the source of truth for phase order and progress. The decisions file records architectural choices and their rationale.

## Phase-based workflow

- Work on one phase at a time.
- Do not start a later phase while an earlier phase has unchecked required work unless the user explicitly approves it.
- Keep each change small enough to review independently.
- Run the tests listed for the current phase before marking work complete.
- Update the checkbox for every completed phase task in `docs/implementation-plan.md`.
- Mark a phase complete only after all of its task checkboxes and review checkpoint are complete.
- If implementation changes an architectural decision, update `docs/decisions.md` in the same change and tell the user.
- Do not mark work complete based only on code being written; verify it with the phase’s acceptance checks.

## Scope boundaries

- Ask before expanding the current phase.
- Keep API schema changes synchronized with the documented dashboard API design.
- Preserve append-only net-worth history.
- Do not use `float64` for financial values or calculations; use `shopspring/decimal`.
- Keep Vite and the Go API as separate processes.
- Name constants using `CAPITAL_SNAKE_CASE` everywhere, for example `DEFAULT_PORT`.
- Name string constants using `CAPITAL_SNAKE_CASE` as well, for example `HEALTH_ROUTE` and `JSON_CONTENT_TYPE`.

## Progress updates

At the end of each implementation session, report:

- Which phase and checkbox items were completed.
- Tests or checks run and their results.
- Any blocked or deferred items.
- Whether the plan or decisions file changed.
