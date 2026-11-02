# syntax=docker/dockerfile:1.7
# ponytail: multi-stage dev/prod; relative VITE_API_BASE_URL eliminates cross-device CORS/CSP rebuilds
FROM node:22-alpine AS base
WORKDIR /app
RUN apk add --no-cache libc6-compat
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# Target: dev (Vite HMR server)
FROM base AS dev
COPY . .
ENV CHOKIDAR_USEPOLLING=true
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# Target: builder (Static SPA export)
FROM base AS builder
ARG VITE_API_BASE_URL=/api/v1
ARG VITE_API_MODE=http
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL} \
    VITE_API_MODE=${VITE_API_MODE}
COPY . .
RUN npm run build

# Target: prod (Nginx server)
FROM nginx:1.27-alpine AS prod
RUN rm -rf /etc/nginx/conf.d/*
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/build /usr/share/nginx/html
EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
