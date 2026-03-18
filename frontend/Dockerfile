FROM oven/bun:1-alpine AS builder
ARG PUBLIC_AUTH_BACKEND=""
ARG PUBLIC_IS_DEV_WORKER=""
ARG PUBLIC_AUTH0_CLIENT_ID=""
ARG PUBLIC_AUTH0_DOMAIN=""
ARG PUBLIC_AUTH0_AUDIENCE=""

COPY ./frontend/ /app/frontend/
COPY ./VERSION /app/
WORKDIR /app/frontend
RUN bun remove @sentry/sveltekit
RUN bun install --frozen-lockfile
WORKDIR /app/frontend/webapp
RUN rm src/instrumentation.server.ts
RUN rm src/hooks.client.ts
RUN rm -r src/routes/sentry
RUN cp vite.config.docker.ts vite.config.ts
RUN cp svelte.config.docker.js svelte.config.js
RUN cp src/hooks.server.docker.ts src/hooks.server.ts
RUN bun remove @sentry/sveltekit
RUN bun install --frozen-lockfile
RUN bun run build

FROM oven/bun:1-distroless AS runner
ENV NODE_ENV=production
COPY --from=builder /app/ /app/
COPY --from=builder /etc/passwd /etc/passwd
EXPOSE 3000
USER nobody
CMD ["node", "/app/frontend/webapp/build/index.js"]
