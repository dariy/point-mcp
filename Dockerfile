FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /point-mcp ./cmd/point-mcp

FROM --platform=$BUILDPLATFORM alpine:3.22 AS certs
RUN apk add --no-cache ca-certificates

FROM alpine:3.22
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /point-mcp /point-mcp
ENTRYPOINT ["/point-mcp"]
