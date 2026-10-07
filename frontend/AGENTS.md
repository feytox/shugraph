context: frontend_monorepo
cwd: frontend

commands:
  pre_commit: pnpm typecheck && pnpm lint && pnpm test
  test_single: pnpm --filter graph-frontapp test <path_to_test>

rules:
  stack_ui: "Mantine v7 only. Tabler icons (@tabler/icons-react). No custom CSS/inline styles unless necessary."
  stack_data: "TanStack Query + @/api/client. Never use raw useEffect/fetch/axios for server state."
  business_logic: "Frontend is strictly a view layer. No client-side risk/graph calculations."
  dates: "Strict UTC via dayjs.utc()"
  dev_env: "Proxy /api -> :8080 (override: VITE_API_PROXY_TARGET) or MSW (VITE_USE_MOCKS=true)"