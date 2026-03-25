# snmp-trap-printer

A CLI tool that listens for SNMP traps (v1, v2c, v3) on a UDP port and prints each trap to stdout as it arrives. Designed for troubleshooting trap pipelines, verifying trap contents, and demoing SNMP integrations without needing a full NMS.

Features:
- Supports SNMPv1, v2c, and v3 (noAuthNoPriv, authNoPriv, authPriv)
- Human-readable labeled key-value output or machine-readable NDJSON
- OID-to-name resolution via standard MIBs (IF-MIB, IP-MIB, SNMPv2-MIB, etc.)
- MIB-aware value formatting: enum names, units, timetick durations, hex for binary data
- Graceful degradation — raw OIDs and values are shown when no MIBs are available

## Build

Requires Go 1.21+.

```bash
go build -o build/snmp-trap-printer ./cmd/snmp-trap-printer/...
```

## Test

```bash
# Unit tests (no network, no root required)
go test ./...

# MIB and formatter tests only
go test ./internal/mib/... ./internal/output/...

# Integration tests (binds real UDP ports, no root required)
go test -tags integration ./internal/trap/...
```

## Run

Port 162 is the IANA-assigned SNMP trap port and requires root (or `CAP_NET_BIND_SERVICE`). Use `--port 10162` during development to avoid elevated privileges.

```bash
# Default: bind 0.0.0.0:162, human output
sudo ./build/snmp-trap-printer

# Development port, human output
./build/snmp-trap-printer --port 10162

# JSON (NDJSON) output
./build/snmp-trap-printer --port 10162 --output json

# Custom bind address
./build/snmp-trap-printer --address 127.0.0.1 --port 10162

# Load extra MIB files from a custom directory
./build/snmp-trap-printer --port 10162 --mib-path /opt/vendor/mibs --mib-path /usr/local/share/snmp/mibs
```

### CLI flags

| Flag | Default | Description |
|------|---------|-------------|
| `--address` | `0.0.0.0` | UDP bind address |
| `--port` | `162` | UDP port |
| `--output` | `human` | `human` or `json` |
| `--mib-path` | _(OS default)_ | Extra MIB directory; repeatable |

### SNMPv3 credentials (environment variables)

| Variable | Description |
|----------|-------------|
| `SNMP_V3_USERNAME` | USM username |
| `SNMP_V3_AUTH_PROTOCOL` | `MD5`, `SHA`, `SHA224`, `SHA256`, `SHA384`, `SHA512` |
| `SNMP_V3_AUTH_PASSWORD` | Authentication passphrase |
| `SNMP_V3_PRIV_PROTOCOL` | `DES`, `AES`, `AES192`, `AES256` |
| `SNMP_V3_PRIV_PASSWORD` | Privacy passphrase |

Security level is inferred automatically:
- No username → noAuthNoPriv
- Username only → noAuthNoPriv
- Username + auth → authNoPriv
- Username + auth + priv → authPriv

Example for authPriv:
```bash
export SNMP_V3_USERNAME=trapuser
export SNMP_V3_AUTH_PROTOCOL=SHA256
export SNMP_V3_AUTH_PASSWORD=authsecret
export SNMP_V3_PRIV_PROTOCOL=AES
export SNMP_V3_PRIV_PASSWORD=privsecret
./build/snmp-trap-printer --port 10162
```

## MIBs

Standard MIBs are loaded automatically from OS default paths if they exist:

| OS | Paths |
|----|-------|
| Linux | `/usr/share/snmp/mibs`, `/usr/local/share/snmp/mibs` |
| macOS | `/usr/local/share/snmp/mibs`, `/opt/homebrew/share/snmp/mibs` |
| Windows | `C:\usr\mibs`, `%APPDATA%\snmp\mibs` |

If no MIBs are found, the tool runs normally and displays raw numeric OIDs and values instead of names.

Install standard MIBs:
```bash
# macOS
brew install net-snmp

# Debian / Ubuntu
sudo apt-get install snmp-mibs-downloader
sudo download-mibs
```

## Docker

A `deploy/docker-compose.yml` is provided for running the container without building manually. The image bundles standard MIBs from `deploy/files/` — host MIB paths are not used.

```bash
# Human output (default), listening on UDP 10162
cd deploy && docker compose up

# JSON (NDJSON) output
cd deploy && OUTPUT_FORMAT=json docker compose up

# SNMPv3 authPriv with JSON output
cd deploy && OUTPUT_FORMAT=json \
  SNMP_V3_USERNAME=trapuser \
  SNMP_V3_AUTH_PROTOCOL=SHA256 \
  SNMP_V3_AUTH_PASSWORD=authsecret \
  SNMP_V3_PRIV_PROTOCOL=AES \
  SNMP_V3_PRIV_PASSWORD=privsecret \
  docker compose up
```

To listen on the privileged port 162 on the host, change the port mapping in `docker-compose.yml` to `"162:10162/udp"` and add `cap_add: [NET_BIND_SERVICE]`.

To load additional MIBs, mount a directory into the container and pass `--mib-path` via the `command` override:

```yaml
services:
  snmp-trap-printer:
    # ...
    volumes:
      - /path/to/your/mibs:/mibs:ro
    command: ["--port", "10162", "--mib-path", "/mibs"]
```

## Sending sample traps

The examples below use `snmptrap` from the [net-snmp](http://www.net-snmp.org/) package.

```bash
# macOS
brew install net-snmp

# Debian / Ubuntu
sudo apt-get install snmp
```

### SNMPv1

```bash
snmptrap -v1 -c public localhost:10162 \
  .1.3.6.1.4.1.9 127.0.0.1 6 1 12345 \
  .1.3.6.1.4.1.9.1.0 i 99
```

Arguments: `<enterprise-OID> <agent-addr> <generic-trap> <specific-trap> <timestamp> [OID type value ...]`

### SNMPv2c — linkDown

```bash
snmptrap -v2c -c public localhost:10162 \
  12345 .1.3.6.1.6.3.1.1.5.3 \
  .1.3.6.1.2.1.2.2.1.1.1 i 1 \
  .1.3.6.1.2.1.2.2.1.7.1 i 1 \
  .1.3.6.1.2.1.2.2.1.8.1 i 2
```

Arguments: `<uptime> <trap-OID> [OID type value ...]`

Type codes: `i` = Integer, `s` = OctetString, `o` = OID, `t` = TimeTicks, `a` = IPAddress, `c` = Counter32, `g` = Gauge32, `C` = Counter64

### SNMPv3 — noAuthNoPriv

```bash
snmptrap -v3 -u trapuser -l noAuthNoPriv \
  localhost:10162 12345 .1.3.6.1.6.3.1.1.5.1 \
  .1.3.6.1.2.1.1.5.0 s "myrouter"
```

### SNMPv3 — authNoPriv

```bash
snmptrap -v3 -u trapuser -l authNoPriv -a MD5 -A authsecret \
  localhost:10162 12345 .1.3.6.1.6.3.1.1.5.1 \
  .1.3.6.1.2.1.1.5.0 s "myrouter"
```

### SNMPv3 — authPriv

```bash
snmptrap -v3 -u trapuser -l authPriv -a SHA -A authsecret -x AES -X privsecret \
  localhost:10162 12345 .1.3.6.1.6.3.1.1.5.1 \
  .1.3.6.1.2.1.1.5.0 s "myrouter"
```

## Output examples

### Human (default)

```
--------------------------------------------------------------------------------
Version:    v2c
Source:     127.0.0.1:55432
Community:  public
Timestamp:  12345 (2m3.45s)
Trap OID:   IF-MIB::linkDown
Varbinds:
  IF-MIB::ifIndex.1                        1
  IF-MIB::ifAdminStatus.1                  up
  IF-MIB::ifOperStatus.1                   down
--------------------------------------------------------------------------------
```

### JSON (`--output json`)

One JSON object per line (NDJSON), suitable for piping to `jq` or ingestion by log collectors.

```json
{"version":"v2c","source_ip":"127.0.0.1:55432","community":"public","timestamp":12345,"oid":"IF-MIB::linkDown","varbinds":[{"oid":".1.3.6.1.2.1.2.2.1.1.1","name":"IF-MIB::ifIndex.1","type":"Integer","value":"1"},{"oid":".1.3.6.1.2.1.2.2.1.7.1","name":"IF-MIB::ifAdminStatus.1","type":"Integer","value":"up"},{"oid":".1.3.6.1.2.1.2.2.1.8.1","name":"IF-MIB::ifOperStatus.1","type":"Integer","value":"down"}]}
```

Filter with `jq`:
```bash
./build/snmp-trap-printer --port 10162 --output json | jq '.varbinds[] | {name, value}'
```
