# Stage 1: Build
FROM golang:1.26.1-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux \
    go build -o /snmp-trap-printer ./cmd/snmp-trap-printer/...

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static-debian12

COPY --from=build /snmp-trap-printer /snmp-trap-printer
COPY deploy/files/ /usr/share/snmp/mibs/

EXPOSE 10162/udp

ENTRYPOINT ["/snmp-trap-printer"]
CMD ["--port", "10162"]
