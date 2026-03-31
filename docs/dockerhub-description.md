A lightweight CLI tool that listens for SNMP traps (v1, v2c, v3) on a UDP port and prints each trap to stdout as it arrives. Designed for troubleshooting trap pipelines, verifying trap contents, and demoing SNMP integrations without needing a full NMS.

## Features

- Supports SNMPv1, v2c, and v3 (noAuthNoPriv, authNoPriv, authPriv)
- Human-readable labeled key-value output or machine-readable NDJSON
- OID-to-name resolution via bundled standard MIBs (IF-MIB, IP-MIB, SNMPv2-MIB, etc.)
- MIB-aware value formatting: enum names, units, timetick durations, hex for binary data
- Mount a directory of custom/vendor MIBs for device-specific OID resolution
- Graceful degradation — raw OIDs and values shown when a MIB is unavailable

## Quick Start

```bash
# Human-readable output on UDP port 10162
docker run --rm -p 10162:10162/udp cblauvelt/snmp-trap-printer:latest

# NDJSON output (pipe to jq, log collectors, etc.)
docker run --rm -p 10162:10162/udp \
  -e OUTPUT_FORMAT=json \
  cblauvelt/snmp-trap-printer:latest
```

## Custom MIB Directory

Mount a host directory and pass `--mib-path` to load vendor or device-specific MIBs:

```bash
docker run --rm -p 10162:10162/udp \
  -v /path/to/your/mibs:/mibs:ro \
  cblauvelt/snmp-trap-printer:latest \
  --port 10162 --mib-path /mibs
```

## SNMPv3

Set credentials via environment variables:

```bash
docker run --rm -p 10162:10162/udp \
  -e SNMP_V3_USERNAME=trapuser \
  -e SNMP_V3_AUTH_PROTOCOL=SHA256 \
  -e SNMP_V3_AUTH_PASSWORD=authsecret \
  -e SNMP_V3_PRIV_PROTOCOL=AES \
  -e SNMP_V3_PRIV_PASSWORD=privsecret \
  cblauvelt/snmp-trap-printer:latest
```

| Variable | Description |
|----------|-------------|
| `SNMP_V3_USERNAME` | USM username |
| `SNMP_V3_AUTH_PROTOCOL` | `MD5`, `SHA`, `SHA224`, `SHA256`, `SHA384`, `SHA512` |
| `SNMP_V3_AUTH_PASSWORD` | Authentication passphrase |
| `SNMP_V3_PRIV_PROTOCOL` | `DES`, `AES`, `AES192`, `AES256` |
| `SNMP_V3_PRIV_PASSWORD` | Privacy passphrase |

Security level is inferred automatically from the variables provided.

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--address` | `0.0.0.0` | UDP bind address |
| `--port` | `162` | UDP port (use 10162 to avoid requiring NET_BIND_SERVICE) |
| `--output` | `human` | `human` or `json` |
| `--mib-path` | _(bundled)_ | Extra MIB directory; repeatable |

## Source

[github.com/cblauvelt/snmp-trap-printer](https://github.com/cblauvelt/snmp-trap-printer)
