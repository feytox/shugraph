FROM node:22-alpine AS builder
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@latest --activate

COPY pnpm-lock.yaml pnpm-workspace.yaml package.json tsconfig.base.json ./
COPY graph-frontapp/package.json ./graph-frontapp/
COPY graph-admin-frontapp/package.json ./graph-admin-frontapp/
RUN pnpm install --frozen-lockfile

COPY api-spec/ ./api-spec/
COPY graph-frontapp/ ./graph-frontapp/
RUN pnpm gen:api && pnpm --filter graph-frontapp build

FROM nginx:alpine AS runner

COPY --from=builder /app/graph-frontapp/dist /usr/share/nginx/html
COPY graph-frontapp/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
