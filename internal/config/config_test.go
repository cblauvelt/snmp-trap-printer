package config

import (
	"os"
	"os/exec"
	"testing"
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
