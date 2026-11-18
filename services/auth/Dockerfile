FROM golang:1.27.4-alpine3.24 AS builder

WORKDIR /app

COPY . .

WORKDIR services/auth

RUN go mod download

RUN mkdir -p ./bin && \
    CGO_ENABLED=0 GOOS=linux \
    go build -o ./bin/auth ./auth/cmd

FROM alpine:3.24 AS runner

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/services/auth/bin/auth ./bin/auth

EXPOSE 8081

CMD ["./bin/auth"]