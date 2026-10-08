# AGENTS.md

> Frontend workspace (`cwd: frontend`).

## Commands

- **Dev**: `pnpm dev` (both apps), `pnpm dev:app` (:5173), `pnpm dev:admin` (:5174)
- **API types**: `pnpm gen:api` -> import types from `@/api/client` (never declare manually)
- **Test single**: `pnpm --filter graph-frontapp test <path_to_test>`
- **Pre-commit**: `pnpm typecheck && pnpm lint && pnpm test`

## Rules

- **UI**: Mantine v9 style props (`c`, `p`, `m`, `fw`) or CSS modules. Zero inline `style={{}}` attributes. Icons from `@tabler/icons-react`.
- **Data**: TanStack Query + `@/api/client`. Never use raw `useEffect`/`fetch`/`axios` for server state.
- **Domain**: Strictly view layer. Zero client-side risk scoring or graph traversals (fetch pre-calculated from API).
- **Dates**: Strict UTC via `dayjs.utc()`.
- **Environment**: Proxy `/api` -> `:8080` (override: `VITE_API_PROXY_TARGET`) or MSW (`VITE_USE_MOCKS=true`).
