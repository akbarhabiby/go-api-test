FROM golang:1.20.4 as BUILD
WORKDIR /app
COPY . .

RUN make build-linux

FROM alpine:3.18.0
WORKDIR /app

COPY --from=BUILD /app/bin/app /app/app

ENTRYPOINT ["/app/app"]
