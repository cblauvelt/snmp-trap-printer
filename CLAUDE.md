# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`snmp-trap-printer` is a Go CLI tool that listens for SNMP traps (v1, v2c, v3) on a UDP port and prints them to stdout for troubleshooting and demo purposes. It supports human-readable labeled KV output and NDJSON output, with OID-to-name resolution via standard MIBs.

## Pull Request Test Plans

When writing the test plan section of a PR:
- Only include tests that are already implemented as unit or integration tests in the codebase.
- Do not add test plan items for tests that do not yet exist.
- Mark every item with `[x]` (completed) after verifying the tests pass — the test plan informs reviewers that these checks have already been run.

## Git Conventions

Commit messages follow the [Conventional Commits](https://www.conventionalcommits.org/) standard:

```
<type>(<scope>): <description>

[optional body]
```

Common types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`.
Scope should reflect the package (e.g., `trap`, `mib`, `output`, `config`).

Examples:
```
feat(trap): add SNMPv1 trap parser
fix(mib): handle missing OID suffix in Translate
refactor(output): extract timetick formatting to shared helper
```

## Branching & Pull Requests

Branch names follow the pattern `{username}/issue{N}` — e.g., `cblauvelt/issue1`. Create one branch per GitHub issue. **Never commit directly to `main`.**

Workflow for each issue:
1. Create branch: `git checkout -b cblauvelt/issue{N}`
2. Commit work on that branch
3. Push and open a PR targeting `main`: `gh pr create --base main`
4. The PR body should include `Closes #N` to auto-close the issue on merge

## Build & Run

```bash
go build -o build/snmp-trap-printer ./cmd/snmp-trap-printer/...
./build/snmp-trap-printer --help

# Run with defaults (UDP 0.0.0.0:162, human output)
sudo ./build/snmp-trap-printer

# Run with options
./build/snmp-trap-printer --port 10162 --output json --mib-path /custom/mibs
```

> Port 162 requires root or `CAP_NET_BIND_SERVICE`. Use `--port 10162` during development to avoid sudo.

## Test

```bash
go test ./...
go test ./internal/mib/...         # MIB-specific tests
go test -run TestParseV1 ./internal/trap/...
```

## Key Dependencies

- `github.com/gosnmp/gosnmp` — SNMP protocol (trap listener, PDU parsing, all versions)
- `github.com/sleepinggenius2/gosmi` — pure-Go MIB parsing (port of libsmi); uses global state, initialize once at startup

## Package Architecture

```
cmd/snmp-trap-printer/main.go   Entry point: wire config → MIB loader → listener → formatter
internal/config/                CLI flags + SNMPv3 env var credentials
internal/trap/                  UDP listener and per-version PDU parsers → shared Trap struct
internal/mib/                   OS path resolution, gosmi initialization, OID translation, value formatting
internal/output/                Human-readable and JSON formatters
```

### Data flow

```
UDP packet
  → gosnmp TrapListener (internal/trap/listener.go)
  → ParseV1 / ParseV2c / ParseV3 (version-specific files)
  → Trap struct (internal/trap/trap.go)
  → mib.Translate(oid) + mib.FormatValue(varbind)
  → HumanFormatter or JSONFormatter → stdout
```

### Shared Trap struct (`internal/trap/trap.go`)

All version parsers produce this struct. Version-specific fields are left zero/empty when not applicable:

```go
type Trap struct {
    Version         gosnmp.SnmpVersion
    SourceIP        string
    Community       string   // v1/v2c
    Enterprise      string   // v1 only
    AgentAddress    string   // v1 only
    GenericType     int      // v1 only
    SpecificType    int      // v1 only
    Timestamp       uint     // sysUpTime timeticks
    OID             string   // v2c/v3 trap OID
    SecurityName    string   // v3 only
    ContextName     string   // v3 only
    ContextEngineID string   // v3 only
    Varbinds        []Varbind
}
```

### Formatter interface (`internal/output/`)

Both formatters implement:
```go
type Formatter interface {
    Format(w io.Writer, trap *trap.Trap) error
}
```

Human output: labeled KV block, 80-char divider lines, one block per trap.
JSON output: one JSON object per line (NDJSON), fields omitted with `omitempty` when not applicable to the SNMP version.

### MIB loading (`internal/mib/`)

gosmi uses global state — `NewLoader` must be called once at startup. The loader:
1. Resolves OS-aware default paths + `--mib-path` flags (`paths.go`)
2. Calls `gosmi.Init()` and loads standard MIBs (`loader.go`)
3. Exposes `Translate(oid string) (name string, found bool)` for OID→name
4. Exposes `FormatValue(loader, varbind)` for type-aware value formatting (`values.go`)

If no MIBs load, the app degrades gracefully — raw OIDs are displayed instead of names.

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--address` | `0.0.0.0` | Bind address |
| `--port` | `162` | UDP port |
| `--output` | `human` | `human` or `json` |
| `--mib-path` | _(OS default)_ | Extra MIB dir; repeatable |

## SNMPv3 Credentials (env vars)

| Variable | Description |
|----------|-------------|
| `SNMP_V3_USERNAME` | USM username |
| `SNMP_V3_AUTH_PROTOCOL` | `MD5`, `SHA`, `SHA224`, `SHA256`, `SHA384`, `SHA512` |
| `SNMP_V3_AUTH_PASSWORD` | Auth passphrase |
| `SNMP_V3_PRIV_PROTOCOL` | `DES`, `AES`, `AES192`, `AES256` |
| `SNMP_V3_PRIV_PASSWORD` | Privacy passphrase |

Security level is inferred: no username → noAuthNoPriv; username only → noAuthNoPriv; + auth → authNoPriv; + priv → authPriv.

## OS Default MIB Paths

| OS | Paths checked |
|----|--------------|
| Linux | `/usr/share/snmp/mibs`, `/usr/local/share/snmp/mibs` |
| macOS | `/usr/local/share/snmp/mibs`, `/opt/homebrew/share/snmp/mibs` |
| Windows | `C:\usr\mibs`, `%APPDATA%\snmp\mibs` |

Only paths that exist on disk are used; missing paths are skipped with a debug log.
