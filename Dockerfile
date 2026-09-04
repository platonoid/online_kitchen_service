FROM golang:1.24.4-alpine AS builder
WORKDIR /src

COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/kitchen-api ./cmd/kitchen-api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/kitchen-api /app/kitchen-api
COPY --from=builder /src/migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/kitchen-api"]
