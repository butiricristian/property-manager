# --- Frontend ---
FROM node:24-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# --- Backend + migrate CLI ---
FROM golang:1.25-alpine AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server
RUN CGO_ENABLED=0 go install -tags postgres -ldflags="-s -w" \
    github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3 \
    && cp /go/bin/migrate /out/migrate

# --- Runtime ---
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H app
WORKDIR /app
COPY --from=server /out/server /out/migrate ./
COPY --from=web /src/web/dist ./web/dist
COPY internal/database/migrations ./migrations
RUN printf '#!/bin/sh\nexec ./migrate -path ./migrations -database "$DATABASE_URL" up\n' > release.sh \
    && chmod +x release.sh
USER app
ENV PORT=8080 APP_ENV=production
EXPOSE 8080
CMD ["./server"]
