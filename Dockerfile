FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o app .

FROM alpine:3.21

WORKDIR /root/

COPY --from=builder /app/app .

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip .
ENV ZONEINFO=/root/zoneinfo.zip

CMD ["./app"]
