FROM docker.io/library/golang:1.22-alpine as builder
WORKDIR /app
ENV CGO_ENABLED=0
COPY main.go go.mod ./
RUN go build -ldflags "-s -w" -o pvpc_exporter main.go

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/pvpc_exporter /app/pvpc_exporter
ENTRYPOINT ["/app/pvpc_exporter"]