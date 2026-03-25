package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
	"github.com/gosnmp/gosnmp"
)

// --- helpers ---

func v1Trap() *trap.Trap {
	return &trap.Trap{
		Version:      gosnmp.Version1,
		SourceIP:     "10.0.0.10:55000",
		Community:    "public",
		Enterprise:   ".1.3.6.1.4.1.9",
		AgentAddress: "10.0.0.5",
		GenericType:  6,
		SpecificType: 1,
		Timestamp:    6000,
		Varbinds: []trap.Varbind{
			{OID: ".1.3.6.1.4.1.9.1.0", Name: "ciscoMgmt", Type: gosnmp.Integer, Value: 1},
			{OID: ".1.3.6.1.4.1.9.2.0", Name: "", Type: gosnmp.Integer, Value: 2},
		},
	}
}

func v2cTrap() *trap.Trap {
	return &trap.Trap{
		Version:   gosnmp.Version2c,
		SourceIP:  "10.0.0.20:162",
		Community: "public",
		OID:       ".1.3.6.1.4.1.9.9.1",
		Timestamp: 360000,
		Varbinds: []trap.Varbind{
			{OID: ".1.3.6.1.4.1.9.1.0", Type: gosnmp.Integer, Value: 42},
		},
	}
}

func v3Trap() *trap.Trap {
	return &trap.Trap{
		Version:         gosnmp.Version3,
		SourceIP:        "10.0.0.30:162",
		SecurityName:    "trapuser",
		ContextName:     "mycontext",
		ContextEngineID: "deadbeef",
		OID:             ".1.3.6.1.4.1.9.9.2",
		Timestamp:       100,
	}
}

// --- HumanFormatter ---

func TestHumanFormatter_V1(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	tr := v1Trap()

	err := f.Format(&buf, tr)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}
	out := buf.String()

	mustContain := []string{
		divider,
		"Version:    v1",
		"Source:     10.0.0.10:55000",
		"Community:  public",
		"Enterprise: .1.3.6.1.4.1.9",
		"Agent:      10.0.0.5",
		"Generic:    6",
		"Specific:   1",
		"Timestamp:  6000 timeticks",
		"Varbinds:",
		"ciscoMgmt",     // resolved name used
		".1.3.6.1.4.1.9.2.0", // raw OID for unresolved
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, out)
		}
	}

	mustNotContain := []string{"SecName:", "Context:"}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("output should not contain %q\nfull output:\n%s", s, out)
		}
	}
}

func TestHumanFormatter_V2c(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	tr := v2cTrap()

	f.Format(&buf, tr)
	out := buf.String()

	mustContain := []string{
		"Version:    v2c",
		"Trap OID:   .1.3.6.1.4.1.9.9.1",
		"Community:  public",
		"Timestamp:  360000 timeticks",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, out)
		}
	}

	mustNotContain := []string{"Generic:", "Specific:", "SecName:", "Context:"}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("output should not contain %q\nfull output:\n%s", s, out)
		}
	}
}

func TestHumanFormatter_V3(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	tr := v3Trap()

	f.Format(&buf, tr)
	out := buf.String()

	mustContain := []string{
		"Version:    v3",
		"SecName:    trapuser",
		"Context:    mycontext",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, out)
		}
	}

	if strings.Contains(out, "Community:") {
		t.Errorf("output should not contain Community for v3 trap\nfull output:\n%s", out)
	}
}

func TestHumanFormatter_ZeroTimestamp(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	tr := &trap.Trap{
		Version:  gosnmp.Version2c,
		SourceIP: "10.0.0.1:162",
		// Timestamp = 0 (zero value)
	}
	f.Format(&buf, tr)
	out := buf.String()

	if strings.Contains(out, "Timestamp:") {
		t.Errorf("output should not contain Timestamp when Timestamp == 0\nfull output:\n%s", out)
	}
}

func TestHumanFormatter_VarbindNameFallback(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	tr := &trap.Trap{
		Version:  gosnmp.Version2c,
		SourceIP: "10.0.0.1:162",
		Varbinds: []trap.Varbind{
			{OID: ".1.3.6.1.4.1.9.1.0", Name: "", Value: 1},     // no name — OID should appear
			{OID: ".1.3.6.1.4.1.9.2.0", Name: "resolvedName", Value: 2}, // name set — Name should appear
		},
	}
	f.Format(&buf, tr)
	out := buf.String()

	if !strings.Contains(out, ".1.3.6.1.4.1.9.1.0") {
		t.Errorf("OID should be printed when Name is empty\nfull output:\n%s", out)
	}
	if !strings.Contains(out, "resolvedName") {
		t.Errorf("resolved Name should be printed when set\nfull output:\n%s", out)
	}
	if strings.Contains(out, ".1.3.6.1.4.1.9.2.0") {
		t.Errorf("OID should not be printed when Name is set\nfull output:\n%s", out)
	}
}

func TestHumanFormatter_NoVarbinds(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	tr := &trap.Trap{
		Version:  gosnmp.Version2c,
		SourceIP: "10.0.0.1:162",
		Varbinds: []trap.Varbind{},
	}
	f.Format(&buf, tr)
	out := buf.String()

	if strings.Contains(out, "Varbinds:") {
		t.Errorf("output should not contain Varbinds section when empty\nfull output:\n%s", out)
	}
}

func TestHumanFormatter_ReturnsNilError(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	err := f.Format(&buf, v1Trap())
	if err != nil {
		t.Errorf("Format returned non-nil error: %v", err)
	}
}

func TestHumanFormatter_UsesFormattedValue(t *testing.T) {
	var buf bytes.Buffer
	f := NewHumanFormatter()
	tr := &trap.Trap{
		Version:  gosnmp.Version2c,
		SourceIP: "10.0.0.1:162",
		Varbinds: []trap.Varbind{
			{OID: ".1.3.6", Name: "sysName", FormattedValue: "up(1)", Type: gosnmp.Integer, Value: 1},
			{OID: ".1.3.6.1", Name: "", FormattedValue: "", Type: gosnmp.Integer, Value: 99},
		},
	}
	f.Format(&buf, tr)
	out := buf.String()

	if !strings.Contains(out, "up(1)") {
		t.Errorf("output should use FormattedValue when set, got:\n%s", out)
	}
	if !strings.Contains(out, "99") {
		t.Errorf("output should fall back to Value when FormattedValue is empty, got:\n%s", out)
	}
}

// --- JSONFormatter ---

func TestJSONFormatter_V1(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter()
	tr := v1Trap()

	err := f.Format(&buf, tr)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	assertJSONField(t, m, "version", "v1")
	assertJSONField(t, m, "source_ip", "10.0.0.10:55000")
	assertJSONField(t, m, "community", "public")
	assertJSONField(t, m, "enterprise", ".1.3.6.1.4.1.9")
	assertJSONField(t, m, "agent_address", "10.0.0.5")
	assertJSONField(t, m, "timestamp", float64(6000))

	// These are set to non-zero for v1 trap
	if _, ok := m["generic_type"]; !ok {
		t.Error("generic_type should be present when non-zero")
	}
	if _, ok := m["specific_type"]; !ok {
		t.Error("specific_type should be present when non-zero")
	}

	// v1-only: no oid or security fields
	assertJSONAbsent(t, m, "oid")
	assertJSONAbsent(t, m, "security_name")
	assertJSONAbsent(t, m, "context_name")
	assertJSONAbsent(t, m, "context_engine_id")

	vbs, ok := m["varbinds"].([]interface{})
	if !ok {
		t.Fatal("varbinds should be a JSON array")
	}
	if len(vbs) != 2 {
		t.Errorf("varbinds length = %d, want 2", len(vbs))
	}
}

func TestJSONFormatter_V2c(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter()
	tr := v2cTrap()

	f.Format(&buf, tr)

	var m map[string]interface{}
	json.Unmarshal(buf.Bytes(), &m)

	assertJSONField(t, m, "version", "v2c")
	assertJSONField(t, m, "oid", ".1.3.6.1.4.1.9.9.1")

	// v1-only fields omitted (omitempty, zero values)
	assertJSONAbsent(t, m, "enterprise")
	assertJSONAbsent(t, m, "agent_address")
	assertJSONAbsent(t, m, "generic_type")
	assertJSONAbsent(t, m, "specific_type")
	assertJSONAbsent(t, m, "security_name")
}

func TestJSONFormatter_V3(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter()
	tr := v3Trap()

	f.Format(&buf, tr)

	var m map[string]interface{}
	json.Unmarshal(buf.Bytes(), &m)

	assertJSONField(t, m, "version", "v3")
	assertJSONField(t, m, "security_name", "trapuser")
	assertJSONField(t, m, "context_name", "mycontext")
	assertJSONField(t, m, "context_engine_id", "deadbeef")

	assertJSONAbsent(t, m, "community")
	assertJSONAbsent(t, m, "enterprise")
}

func TestJSONFormatter_NoVarbinds(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter()
	tr := &trap.Trap{
		Version:  gosnmp.Version2c,
		SourceIP: "10.0.0.1:162",
		Varbinds: nil,
	}

	f.Format(&buf, tr)

	var m map[string]interface{}
	json.Unmarshal(buf.Bytes(), &m)

	assertJSONAbsent(t, m, "varbinds")
}

func TestJSONFormatter_IsNDJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter()
	f.Format(&buf, v1Trap())

	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		t.Error("NDJSON output must end with a newline")
	}
	// No embedded newlines before the final one
	trimmed := strings.TrimSuffix(out, "\n")
	if strings.Contains(trimmed, "\n") {
		t.Errorf("NDJSON output should be a single line, got multiple lines:\n%s", out)
	}
}

func TestJSONFormatter_ReturnsNilError(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter()
	err := f.Format(&buf, v1Trap())
	if err != nil {
		t.Errorf("Format returned non-nil error: %v", err)
	}
}

func TestJSONFormatter_UsesFormattedValue(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter()
	tr := &trap.Trap{
		Version:  gosnmp.Version2c,
		SourceIP: "10.0.0.1:162",
		Varbinds: []trap.Varbind{
			{OID: ".1.3.6", Name: "ifOperStatus", FormattedValue: "up", Type: gosnmp.Integer, Value: 1},
			{OID: ".1.3.6.1", Name: "", FormattedValue: "", Type: gosnmp.Integer, Value: 99},
		},
	}
	f.Format(&buf, tr)

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	vbs, _ := m["varbinds"].([]interface{})
	if len(vbs) != 2 {
		t.Fatalf("expected 2 varbinds, got %d", len(vbs))
	}

	vb0, _ := vbs[0].(map[string]interface{})
	if vb0["value"] != "up" {
		t.Errorf("varbind[0].value = %v, want %q", vb0["value"], "up")
	}

	vb1, _ := vbs[1].(map[string]interface{})
	if vb1["value"] != "99" {
		t.Errorf("varbind[1].value = %v, want %q", vb1["value"], "99")
	}
}

// --- versionString ---

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  gosnmp.SnmpVersion
		expected string
	}{
		{gosnmp.Version1, "v1"},
		{gosnmp.Version2c, "v2c"},
		{gosnmp.Version3, "v3"},
		{gosnmp.SnmpVersion(99), "unknown(99)"},
	}
	for _, tt := range tests {
		got := versionString(tt.version)
		if got != tt.expected {
			t.Errorf("versionString(%v) = %q, want %q", tt.version, got, tt.expected)
		}
	}
}

// --- berTypeName ---

func TestBerTypeName(t *testing.T) {
	tests := []struct {
		typ      gosnmp.Asn1BER
		expected string
	}{
		{gosnmp.Integer, "Integer"},
		{gosnmp.OctetString, "OctetString"},
		{gosnmp.ObjectIdentifier, "ObjectIdentifier"},
		{gosnmp.IPAddress, "IPAddress"},
		{gosnmp.Counter32, "Counter32"},
		{gosnmp.Gauge32, "Gauge32"},
		{gosnmp.TimeTicks, "TimeTicks"},
		{gosnmp.Counter64, "Counter64"},
		{gosnmp.Asn1BER(0xFF), "Unknown"},
	}
	for _, tt := range tests {
		got := berTypeName(tt.typ)
		if got != tt.expected {
			t.Errorf("berTypeName(%v) = %q, want %q", tt.typ, got, tt.expected)
		}
	}
}

// --- test helpers ---

func assertJSONField(t *testing.T, m map[string]interface{}, key string, want interface{}) {
	t.Helper()
	got, ok := m[key]
	if !ok {
		t.Errorf("JSON missing field %q", key)
		return
	}
	if got != want {
		t.Errorf("JSON field %q = %v (%T), want %v (%T)", key, got, got, want, want)
	}
}

func assertJSONAbsent(t *testing.T, m map[string]interface{}, key string) {
	t.Helper()
	if _, ok := m[key]; ok {
		t.Errorf("JSON field %q should be absent (omitempty), but present with value %v", key, m[key])
	}
}
