FROM node:22-alpine AS builder
WORKDIR /app
RUN corepack enable

COPY pnpm-lock.yaml pnpm-workspace.yaml package.json tsconfig.base.json ./
COPY graph-frontapp/package.json ./graph-frontapp/
COPY graph-admin-frontapp/package.json ./graph-admin-frontapp/
RUN pnpm install --frozen-lockfile

COPY graph-frontapp/ ./graph-frontapp/
RUN pnpm --filter graph-frontapp build

FROM nginx:alpine AS runner

ENV API_UPSTREAM=http://graph-api:8080/api/

COPY --from=builder /app/graph-frontapp/dist /usr/share/nginx/html
COPY graph-frontapp/default.conf.template /etc/nginx/templates/default.conf.template

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
