FROM node:22-alpine AS builder
WORKDIR /app
RUN corepack enable

COPY pnpm-lock.yaml pnpm-workspace.yaml package.json tsconfig.base.json ./
COPY graph-frontapp/package.json ./graph-frontapp/
COPY graph-admin-frontapp/package.json ./graph-admin-frontapp/
RUN pnpm install --frozen-lockfile

COPY graph-admin-frontapp/ ./graph-admin-frontapp/
RUN pnpm --filter graph-admin-frontapp build

FROM nginxinc/nginx-unprivileged:alpine AS runner

COPY --from=builder /app/graph-admin-frontapp/dist /usr/share/nginx/html/admin
COPY graph-admin-frontapp/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
