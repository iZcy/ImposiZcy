FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

ARG GITHUB_TOKEN
RUN if [ -n "$GITHUB_TOKEN" ]; then \
      git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi
ENV GOPRIVATE="github.com/KreaZcy/*,github.com/iZcy/*"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o imposizcy-server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata chromium
WORKDIR /app

ENV TZ=Asia/Jakarta
ENV GIN_MODE=release
ENV CHROME_PATH=/usr/bin/chromium-browser

COPY --from=builder /app/imposizcy-server .
COPY templates/ ./templates/
COPY public/ ./public/

RUN chmod +x ./imposizcy-server

EXPOSE 9104

CMD ["./imposizcy-server"]
