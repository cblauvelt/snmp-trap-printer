package config

import (
	"os"
	"os/exec"
	"testing"

	"github.com/gosnmp/gosnmp"
)

func TestMibPathList(t *testing.T) {
	var m mibPathList

	if m.String() != "" {
		t.Fatalf("expected empty string, got %q", m.String())
	}

	if err := m.Set("/tmp/mibs1"); err != nil {
		t.Fatal(err)
	}
	if err := m.Set("/tmp/mibs2"); err != nil {
		t.Fatal(err)
	}

	if got := m.String(); got != "/tmp/mibs1,/tmp/mibs2" {
		t.Fatalf("expected /tmp/mibs1,/tmp/mibs2, got %q", got)
	}
	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
}

// TestParseDefaults verifies default flag values by re-executing the test binary
// as a subprocess with no flags set.
func TestParseDefaults(t *testing.T) {
	if os.Getenv("TEST_PARSE_DEFAULTS") == "1" {
		cfg := Parse()
		if cfg.Address != "0.0.0.0" {
			t.Fatalf("address: got %q, want 0.0.0.0", cfg.Address)
		}
		if cfg.Port != 162 {
			t.Fatalf("port: got %d, want 162", cfg.Port)
		}
		if cfg.Output != OutputHuman {
			t.Fatalf("output: got %q, want %q", cfg.Output, OutputHuman)
		}
		if len(cfg.MIBPaths) != 0 {
			t.Fatalf("mib-paths: expected empty, got %v", cfg.MIBPaths)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseDefaults")
	cmd.Env = append(os.Environ(), "TEST_PARSE_DEFAULTS=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

// TestParseOutputJSON verifies that --output json is accepted.
func TestParseOutputJSON(t *testing.T) {
	if os.Getenv("TEST_PARSE_OUTPUT_JSON") == "1" {
		os.Args = []string{"snmp-trap-printer", "--output", "json"}
		cfg := Parse()
		if cfg.Output != OutputJSON {
			t.Fatalf("output: got %q, want %q", cfg.Output, OutputJSON)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseOutputJSON")
	cmd.Env = append(os.Environ(), "TEST_PARSE_OUTPUT_JSON=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

// TestParseOutputInvalid verifies that an invalid --output value exits with code 2.
func TestParseOutputInvalid(t *testing.T) {
	if os.Getenv("TEST_PARSE_OUTPUT_INVALID") == "1" {
		os.Args = []string{"snmp-trap-printer", "--output", "badformat"}
		Parse()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseOutputInvalid")
	cmd.Env = append(os.Environ(), "TEST_PARSE_OUTPUT_INVALID=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit, got nil")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", exitErr.ExitCode())
	}
}

// TestParseOutputFormatEnvVar verifies that OUTPUT_FORMAT env var sets the output format.
func TestParseOutputFormatEnvVar(t *testing.T) {
	if os.Getenv("TEST_PARSE_OUTPUT_FORMAT_ENV") == "1" {
		cfg := Parse()
		if cfg.Output != OutputJSON {
			t.Fatalf("output: got %q, want %q", cfg.Output, OutputJSON)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseOutputFormatEnvVar")
	cmd.Env = append(os.Environ(), "TEST_PARSE_OUTPUT_FORMAT_ENV=1", "OUTPUT_FORMAT=json")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

// TestParseFlagOverridesOutputFormatEnvVar verifies that --output flag takes precedence over OUTPUT_FORMAT env var.
func TestParseFlagOverridesOutputFormatEnvVar(t *testing.T) {
	if os.Getenv("TEST_PARSE_FLAG_OVERRIDES_ENV") == "1" {
		os.Args = []string{"snmp-trap-printer", "--output", "human"}
		cfg := Parse()
		if cfg.Output != OutputHuman {
			t.Fatalf("output: got %q, want %q", cfg.Output, OutputHuman)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseFlagOverridesOutputFormatEnvVar")
	cmd.Env = append(os.Environ(), "TEST_PARSE_FLAG_OVERRIDES_ENV=1", "OUTPUT_FORMAT=json")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

// TestParseMIBPaths verifies that --mib-path can be specified multiple times.
func TestParseMIBPaths(t *testing.T) {
	if os.Getenv("TEST_PARSE_MIB_PATHS") == "1" {
		os.Args = []string{"snmp-trap-printer", "--mib-path", "/tmp/a", "--mib-path", "/tmp/b"}
		cfg := Parse()
		if len(cfg.MIBPaths) != 2 {
			t.Fatalf("expected 2 mib paths, got %d: %v", len(cfg.MIBPaths), cfg.MIBPaths)
		}
		if cfg.MIBPaths[0] != "/tmp/a" || cfg.MIBPaths[1] != "/tmp/b" {
			t.Fatalf("unexpected mib paths: %v", cfg.MIBPaths)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseMIBPaths")
	cmd.Env = append(os.Environ(), "TEST_PARSE_MIB_PATHS=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

func TestLoadV3ConfigNoUsername(t *testing.T) {
	cfg, err := LoadV3Config()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Username != "" {
		t.Fatalf("expected empty username, got %q", cfg.Username)
	}
	if got := cfg.SecurityLevel(); got != gosnmp.NoAuthNoPriv {
		t.Fatalf("security level: got %v, want NoAuthNoPriv", got)
	}
}

func TestLoadV3ConfigUsernameOnly(t *testing.T) {
	t.Setenv("SNMP_V3_USERNAME", "trapuser")

	cfg, err := LoadV3Config()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Username != "trapuser" {
		t.Fatalf("username: got %q, want trapuser", cfg.Username)
	}
	if got := cfg.SecurityLevel(); got != gosnmp.NoAuthNoPriv {
		t.Fatalf("security level: got %v, want NoAuthNoPriv", got)
	}
}

func TestLoadV3ConfigAuthNoPriv(t *testing.T) {
	t.Setenv("SNMP_V3_USERNAME", "trapuser")
	t.Setenv("SNMP_V3_AUTH_PROTOCOL", "SHA256")
	t.Setenv("SNMP_V3_AUTH_PASSWORD", "authpass123")

	cfg, err := LoadV3Config()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AuthProtocol != gosnmp.SHA256 {
		t.Fatalf("auth protocol: got %v, want SHA256", cfg.AuthProtocol)
	}
	if got := cfg.SecurityLevel(); got != gosnmp.AuthNoPriv {
		t.Fatalf("security level: got %v, want AuthNoPriv", got)
	}
}

func TestLoadV3ConfigAuthPriv(t *testing.T) {
	t.Setenv("SNMP_V3_USERNAME", "trapuser")
	t.Setenv("SNMP_V3_AUTH_PROTOCOL", "SHA256")
	t.Setenv("SNMP_V3_AUTH_PASSWORD", "authpass123")
	t.Setenv("SNMP_V3_PRIV_PROTOCOL", "AES")
	t.Setenv("SNMP_V3_PRIV_PASSWORD", "privpass123")

	cfg, err := LoadV3Config()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AuthProtocol != gosnmp.SHA256 {
		t.Fatalf("auth protocol: got %v, want SHA256", cfg.AuthProtocol)
	}
	if cfg.PrivProtocol != gosnmp.AES {
		t.Fatalf("priv protocol: got %v, want AES", cfg.PrivProtocol)
	}
	if got := cfg.SecurityLevel(); got != gosnmp.AuthPriv {
		t.Fatalf("security level: got %v, want AuthPriv", got)
	}
}

func TestLoadV3ConfigInvalidAuthProtocol(t *testing.T) {
	t.Setenv("SNMP_V3_USERNAME", "trapuser")
	t.Setenv("SNMP_V3_AUTH_PROTOCOL", "BADVAL")

	cfg, err := LoadV3Config()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AuthProtocol != gosnmp.NoAuth {
		t.Fatalf("auth protocol: got %v, want NoAuth", cfg.AuthProtocol)
	}
	if got := cfg.SecurityLevel(); got != gosnmp.NoAuthNoPriv {
		t.Fatalf("security level: got %v, want NoAuthNoPriv", got)
	}
}

func TestLoadV3ConfigInvalidPrivProtocol(t *testing.T) {
	t.Setenv("SNMP_V3_USERNAME", "trapuser")
	t.Setenv("SNMP_V3_AUTH_PROTOCOL", "SHA")
	t.Setenv("SNMP_V3_AUTH_PASSWORD", "authpass123")
	t.Setenv("SNMP_V3_PRIV_PROTOCOL", "BADVAL")

	cfg, err := LoadV3Config()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PrivProtocol != gosnmp.NoPriv {
		t.Fatalf("priv protocol: got %v, want NoPriv", cfg.PrivProtocol)
	}
	if got := cfg.SecurityLevel(); got != gosnmp.AuthNoPriv {
		t.Fatalf("security level: got %v, want AuthNoPriv", got)
	}
}

func TestV3ConfigSecurityLevel(t *testing.T) {
	cases := []struct {
		name     string
		cfg      V3Config
		expected gosnmp.SnmpV3MsgFlags
	}{
		{"empty", V3Config{}, gosnmp.NoAuthNoPriv},
		{"username only", V3Config{Username: "u"}, gosnmp.NoAuthNoPriv},
		{"auth no priv", V3Config{Username: "u", AuthProtocol: gosnmp.SHA256, AuthPassword: "p"}, gosnmp.AuthNoPriv},
		{"auth priv", V3Config{Username: "u", AuthProtocol: gosnmp.SHA256, AuthPassword: "p", PrivProtocol: gosnmp.AES, PrivPassword: "q"}, gosnmp.AuthPriv},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.SecurityLevel(); got != tc.expected {
				t.Fatalf("got %v, want %v", got, tc.expected)
			}
		})
	}
}
