# syntax=docker/dockerfile:1

FROM golang:1.26.4-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && adduser -D -H -u 10001 appuser
WORKDIR /app
COPY --from=builder /out/api /app/api
USER appuser
EXPOSE 8080
ENV ADDR=:8080
ENTRYPOINT ["/app/api"]
