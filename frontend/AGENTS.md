# AGENTS.md

> Frontend workspace (`cwd: frontend`). Strictly view layer, no client-side calculations.

## Commands
- **Dev**: `pnpm dev` (app :5173), `pnpm dev:admin` (admin :5174)
- **API types**: `pnpm gen:api` (never write API types manually)
- **Test single**: `pnpm --filter graph-frontapp test <path_to_test>`
- **Pre-commit**: `pnpm typecheck && pnpm lint && pnpm test`

## Rules
- **UI**: Mantine v7 only + `@tabler/icons-react`. No custom CSS/inline styles unless necessary.
- **Data**: TanStack Query + `@/api/client`. Never use raw `useEffect`/`fetch`/`axios` for server state.
- **Dates**: Strict UTC via `dayjs.utc()`.
- **Environment**: Proxy `/api` -> `:8080` (override: `VITE_API_PROXY_TARGET`) or MSW (`VITE_USE_MOCKS=true`).
- **Errors**: Show backend `detail` text to user, not raw HTTP codes.