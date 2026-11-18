FROM golang:1.27.4-alpine3.24 AS builder

WORKDIR /app

COPY . .

WORKDIR services/users

RUN go mod download

RUN mkdir -p ./bin && \
    CGO_ENABLED=0 GOOS=linux \
    go build -o ./bin/users ./users/cmd

FROM alpine:3.24 AS runner

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/services/users/bin/users ./bin/users

EXPOSE 8082

CMD ["./bin/users"]