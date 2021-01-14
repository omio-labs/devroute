FROM golang:1.13-alpine AS builder

WORKDIR /build

COPY . .
RUN go build -o devrouter

FROM alpine:latest

USER nobody
WORKDIR /go/src/app

COPY --from=builder /build/devrouter .
CMD ["./devrouter"]
