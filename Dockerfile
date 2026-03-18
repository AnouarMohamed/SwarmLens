# ── Stage 1: Build Go backend ────────────────────────────────────────────────
FROM golang:1.23-alpine AS backend-builder
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o swarmlens ./cmd/swarmlens

# ── Stage 2: Build frontend ───────────────────────────────────────────────────
FROM node:22-alpine AS frontend-builder
WORKDIR /build
COPY package.json package-lock.json ./
RUN npm ci
COPY index.html vite.config.ts tsconfig.json ./
COPY src/ ./src/
RUN npm run build

# ── Stage 3: Final image ──────────────────────────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /build/swarmlens ./swarmlens

# Copy frontend dist — served by the Go backend as static files
COPY --from=frontend-builder /build/dist ./static

# Non-root user
RUN addgroup -g 1001 swarmlens && adduser -u 1001 -G swarmlens -D swarmlens
USER swarmlens

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v1/healthz || exit 1

ENTRYPOINT ["./swarmlens"]
