FROM golang:1.26-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o imposizcy-server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata chromium

WORKDIR /app

ENV TZ=Asia/Jakarta
ENV GIN_MODE=release
ENV CHROME_PATH=/usr/bin/chromium-browser

COPY --from=builder /build/imposizcy-server .
COPY --from=builder /build/templates/ ./templates/
COPY --from=builder /build/public/ ./public/

RUN chmod +x ./imposizcy-server

EXPOSE 7231

CMD ["./imposizcy-server"]
