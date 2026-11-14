# AGENTS.md

> Frontend workspace (`cwd: frontend`).

## Commands

- **API types**: `pnpm gen:api` -> import types from `@/api/client` (never declare manually)
- **Pre-commit**: `pnpm typecheck && pnpm lint && pnpm test`

## Rules

- **UI**: Mantine v9 style props (`c`, `p`, `m`, `fw`) or CSS modules. Zero inline `style={{}}` attributes. Icons from `@tabler/icons-react`. No hover tooltips (`Tooltip`, `title`); icon-only controls get `aria-label`.
- **Data**: TanStack Query + `@/api/client`. Never use raw `useEffect`/`fetch`/`axios` for server state.
- **Domain**: Strictly view layer. The client may use edges and their timestamps to place nodes on screen, but doesn't derive or show conclusions from them (risk, membership, roles).
- **Structure**: one folder per product feature in `src/features/`. Features don't import each other, `App.tsx` composes them. Code used by more than one feature lives in `src/shared/`. Features exchange state through URL search params and the TanStack Query cache (shared query keys, invalidation after mutations).
- **Dates**: Strict UTC via `dayjs.utc()`.
