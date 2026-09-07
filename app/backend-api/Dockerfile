# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN GOPROXY=https://proxy.golang.org,https://goproxy.io,direct GONOSUMCHECK=* go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags go_json -ldflags="-s -w" -o backend ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o migrate ./cmd/migrate

# --- Runtime Stage ---
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/backend /backend
COPY --from=builder /app/migrate /migrate
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/backend"]