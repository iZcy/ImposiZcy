FROM golang:1.26-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

COPY libs/kzcy-config/go.mod /build/libs/kzcy-config/go.mod
COPY libs/kzcy-dashboard/go.mod /build/libs/kzcy-dashboard/go.mod
COPY services/ImposiZcy/ImposiZcy/go.mod /build/services/ImposiZcy/ImposiZcy/go.mod
COPY services/ImposiZcy/ImposiZcy/go.sum /build/services/ImposiZcy/ImposiZcy/go.sum

COPY go.work /build/go.work

WORKDIR /build/services/ImposiZcy/ImposiZcy

RUN go mod download

COPY libs/kzcy-config/ /build/libs/kzcy-config/
COPY libs/kzcy-dashboard/ /build/libs/kzcy-dashboard/
COPY services/ImposiZcy/ImposiZcy/ /build/services/ImposiZcy/ImposiZcy/

RUN CGO_ENABLED=0 GOOS=linux go build -o imposizcy-server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata chromium

WORKDIR /app

ENV TZ=Asia/Jakarta
ENV GIN_MODE=release
ENV CHROME_PATH=/usr/bin/chromium-browser

COPY --from=builder /build/services/ImposiZcy/ImposiZcy/imposizcy-server .
COPY --from=builder /build/services/ImposiZcy/ImposiZcy/templates/ ./templates/
COPY --from=builder /build/services/ImposiZcy/ImposiZcy/public/ ./public/

RUN chmod +x ./imposizcy-server

EXPOSE 9104

CMD ["./imposizcy-server"]
