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

## Autonomous phase orchestration

When the user asks to implement a phase, act as the manager unless they explicitly ask for direct implementation:

1. Read `AGENTS.md`, `docs/implementation-plan.md`, `docs/decisions.md`, and relevant README files.
2. Identify the current incomplete phase and confirm its scope and acceptance criteria.
3. Create or use a phase-specific branch from the latest base branch.
4. Delegate all code changes to an implementer agent. The manager must not implement, edit, or patch code during this workflow.
5. Delegate a read-only review to a reviewer agent after implementation.
6. Require the reviewer to compare the branch with the base branch and categorize findings as `Critical`, `Warning`, or `Suggestion`.
7. Send every `Critical` and `Warning` finding to the implementer for correction.
8. Repeat the implementer-fix and reviewer-review cycle until no `Critical` or `Warning` findings remain.
9. Run every formatter, test, linter, build, and acceptance check listed for the phase.
10. Update phase checkboxes only after implementation, review, and verification pass. Update `docs/decisions.md` when an architectural decision changes.
11. Commit and push the phase branch only after all required work is complete.
12. Create a draft pull request targeting the base branch and report its URL in the parent chat.

The manager must send concise milestone updates when scope is confirmed, implementation starts or finishes, review findings arrive, fixes are applied, verification passes, the branch is pushed, and the PR is created. The final report must include the phase, checklist items, changed files, review findings and resolutions, checks run, commit SHA, branch, PR URL, and deferred suggestions.

If nested agent delegation is unavailable, create the implementer and reviewer as direct delegated agents and coordinate them from the manager task. Never silently implement or review the work yourself. If a required agent cannot be called, report the blocker rather than claiming completion.

### Pull request description requirements

- Use readable Markdown sections such as `Summary`, `Changes`, `Checks`, and `Deferred Work`.
- Create the description with actual line breaks, preferably in a body file passed through `gh pr create --body-file`.
- Never pass literal `\\n` sequences as PR description content.
- After creation, inspect the body with `gh pr view` and repair malformed Markdown with `gh pr edit --body-file`.
- Relay the complete verified PR URL to the user, including when the PR is a draft.

## Scope boundaries

- Ask before expanding the current phase.
- Keep API schema changes synchronized with the documented dashboard API design.
- Preserve append-only net-worth history.
- Do not use `float64` for financial values or calculations; use `shopspring/decimal`.
- Keep Vite and the Go API as separate processes.
- Name constants using `CAPITAL_SNAKE_CASE` everywhere, for example `DEFAULT_PORT`.
- Name string constants using `CAPITAL_SNAKE_CASE` as well, for example `HEALTH_ROUTE` and `JSON_CONTENT_TYPE`.
- Put `if err != nil` checks on the line immediately after the statement that produces `err`; do not combine the statement and check on one line.
- Keep files focused and reasonably small; split code into multiple files when a file contains unrelated responsibilities or becomes difficult to navigate.

## Progress updates

At the end of each implementation session, report:

- Which phase and checkbox items were completed.
- Tests or checks run and their results.
- Any blocked or deferred items.
- Whether the plan or decisions file changed.
