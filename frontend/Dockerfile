# syntax=docker/dockerfile:1.7
#
# Production image for @els/main-app (pnpm monorepo, Vite build → Nginx).
# Main app imports feature-apps (admin-app, calendar-app, profile-app, project-app,
# timesheets-app) and shared packages (ui, api-client, ...) via workspace:* links,
# so the whole `apps/` + `packages/` tree must be in the build context.

ARG NODE_VERSION=22
ARG NGINX_VERSION=1.29
ARG PNPM_VERSION=10.17.1

# ---- deps ------------------------------------------------------------------
FROM node:${NODE_VERSION}-alpine AS deps

RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate

WORKDIR /src

# Copy only manifests first for a cachable install layer.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY tsconfig.base.json ./
COPY apps ./apps
COPY packages ./packages

# Drop nested source to keep the install layer independent of code changes.
RUN find apps packages -mindepth 2 -maxdepth 2 \
      -not -name 'package.json' -not -name '.' \
      -exec rm -rf {} + 2>/dev/null || true

RUN --mount=type=cache,id=pnpm-store,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

# ---- builder ---------------------------------------------------------------
FROM node:${NODE_VERSION}-alpine AS builder

RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate

WORKDIR /src

COPY --from=deps /src/node_modules ./node_modules
COPY --from=deps /src/apps ./apps
COPY --from=deps /src/packages ./packages

# Bring in the real source tree.
COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ENV VITE_APP_VERSION=${VERSION}
ENV VITE_GIT_COMMIT=${GIT_COMMIT}

RUN --mount=type=cache,id=pnpm-store,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --offline

RUN pnpm --filter @els/main-app build

# ---- runtime ---------------------------------------------------------------
FROM nginx:${NGINX_VERSION}-alpine

RUN apk upgrade --no-cache && \
    rm -rf /usr/share/nginx/html/* /etc/nginx/conf.d/default.conf

COPY --from=builder /src/apps/main-app/dist /usr/share/nginx/html
COPY nginx.conf                              /etc/nginx/conf.d/default.conf

RUN mkdir -p /var/cache/nginx /var/run /tmp && \
    chown -R nginx:nginx /var/cache/nginx /var/run /usr/share/nginx/html /tmp

USER nginx

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

CMD ["nginx", "-g", "daemon off;"]
