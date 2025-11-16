# syntax=docker/dockerfile:1.6

ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/reviewer-service ./cmd/reviewer-service

FROM gcr.io/distroless/static-debian12 AS runner
WORKDIR /app
COPY --from=builder /out/reviewer-service ./reviewer-service
EXPOSE 8080
ENTRYPOINT ["/app/reviewer-service"]
