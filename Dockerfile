FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /cpu-monitor .

FROM scratch

COPY --from=builder /cpu-monitor /cpu-monitor

USER 65532:65532
ENTRYPOINT ["/cpu-monitor"]
