FROM docker.io/library/golang:alpine AS builder
WORKDIR /app
ENV CGO_ENABLED=0
COPY main.go go.mod ./
RUN addgroup -g 10001 user \
    && adduser -H -D -u 10000 -G user user
RUN apk add --quiet --no-cache tzdata && go build -ldflags "-s -w" -trimpath -o pvpc_exporter main.go

FROM scratch
WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/pvpc_exporter /app/pvpc_exporter
USER user:user
ENTRYPOINT ["/app/pvpc_exporter"]