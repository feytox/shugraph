# AGENTS.md

> Frontend workspace (`cwd: frontend`).

## Commands

- **API types**: `pnpm gen:api` -> import types from `@/api/client` (never declare manually)
- **Pre-commit**: `pnpm typecheck && pnpm lint && pnpm test`

## Rules

- **UI**: Mantine v9 style props (`c`, `p`, `m`, `fw`) or CSS modules. Zero inline `style={{}}` attributes. Icons from `@tabler/icons-react`. No hover tooltips (`Tooltip`, `title`); icon-only controls get `aria-label`.
- **Data**: TanStack Query + `@/api/client`. Never use raw `useEffect`/`fetch`/`axios` for server state.
- **Domain**: Strictly view layer. Risk, group membership and node roles come from the API; the client may only look at a node's own edges to place it on screen.
- **Structure**: one folder per product feature in `src/features/`. Features don't import each other, `App.tsx` composes them. Code used by more than one feature lives in `src/shared/`. Features exchange state through URL search params and the TanStack Query cache (shared query keys, invalidation after mutations).
- **Dates**: Strict UTC via `dayjs.utc()`.
- **Environment**: Proxy `/api` -> `:8080` (override: `VITE_API_PROXY_TARGET`) or MSW (`VITE_USE_MOCKS=true`).
