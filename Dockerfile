FROM golang:bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /linxo-reader ./cmd/linxo-reader

FROM alpine:3.20

RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont

ENV CHROME_BIN=/usr/bin/chromium
ENV ROD_BROWSER_PATH=/usr/bin/chromium

WORKDIR /app
COPY --from=builder /linxo-reader .

EXPOSE 8080

ENTRYPOINT ["./linxo-reader"]
