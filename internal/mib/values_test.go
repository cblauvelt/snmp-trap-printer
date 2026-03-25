package mib

import (
	"fmt"
	"math"
	"testing"

	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
	"github.com/gosnmp/gosnmp"
)

// --- formatTimeTicks ---

func TestFormatTimeTicks(t *testing.T) {
	tests := []struct {
		ticks    uint32
		expected string
	}{
		{0, "0 (0s)"},
		{100, "100 (1s)"},
		{99, "99 (990ms)"},
		{6000, "6000 (1m0s)"},
		{6001, "6001 (1m0.01s)"},
		{360000, "360000 (1h0m0s)"},
		{12345, "12345 (2m3.45s)"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("ticks=%d", tt.ticks), func(t *testing.T) {
			got := formatTimeTicks(tt.ticks)
			if got != tt.expected {
				t.Errorf("formatTimeTicks(%d) = %q, want %q", tt.ticks, got, tt.expected)
			}
		})
	}
}

// --- printableOrHex ---

func TestPrintableOrHex(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantHex  bool
		expected string
	}{
		{"all printable", []byte("hello"), false, "hello"},
		{"hello world", []byte("Hello, World!"), false, "Hello, World!"},
		{"empty", []byte{}, false, ""},
		{"space boundary (0x20)", []byte{0x20}, false, " "},
		{"tilde boundary (0x7e)", []byte{0x7e}, false, "~"},
		{"below boundary (0x1f)", []byte{0x1f}, true, ""},
		{"above boundary (0x7f)", []byte{0x7f}, true, ""},
		{"null byte", []byte("abc\x00def"), true, ""},
		{"mixed printable and non-printable", []byte{0x41, 0x00, 0x42}, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := printableOrHex(tt.input)
			if tt.wantHex {
				if len(got) < 2 || got[:2] != "0x" {
					t.Errorf("printableOrHex(%v) = %q, want hex string starting with '0x'", tt.input, got)
				}
			} else {
				if got != tt.expected {
					t.Errorf("printableOrHex(%v) = %q, want %q", tt.input, got, tt.expected)
				}
			}
		})
	}
}

// --- toUint32 ---

func TestToUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantOK    bool
		wantValue uint32
	}{
		{"uint32 zero", uint32(0), true, 0},
		{"uint32 value", uint32(42), true, 42},
		{"uint32 max", uint32(math.MaxUint32), true, math.MaxUint32},
		{"uint value", uint(100), true, 100},
		{"int value", int(200), true, 200},
		{"string", "42", false, 0},
		{"float64", float64(3.14), false, 0},
		{"nil", nil, false, 0},
		{"bool", true, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toUint32(tt.input)
			if ok != tt.wantOK {
				t.Errorf("toUint32(%v) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && got != tt.wantValue {
				t.Errorf("toUint32(%v) = %d, want %d", tt.input, got, tt.wantValue)
			}
		})
	}
}

// --- toInt64 ---

func TestToInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantOK    bool
		wantValue int64
	}{
		{"int", int(42), true, 42},
		{"int32", int32(-1), true, -1},
		{"int64", int64(100), true, 100},
		{"uint", uint(200), true, 200},
		{"uint32", uint32(300), true, 300},
		{"string", "42", false, 0},
		{"float64", float64(3.14), false, 0},
		{"nil", nil, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toInt64(tt.input)
			if ok != tt.wantOK {
				t.Errorf("toInt64(%v) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && got != tt.wantValue {
				t.Errorf("toInt64(%v) = %d, want %d", tt.input, got, tt.wantValue)
			}
		})
	}
}

// --- FormatValue (nil loader — no MIB enrichment) ---

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		vb       trap.Varbind
		expected string
	}{
		// TimeTicks
		{
			"TimeTicks zero",
			trap.Varbind{Type: gosnmp.TimeTicks, Value: uint32(0)},
			"0 (0s)",
		},
		{
			"TimeTicks 1 second",
			trap.Varbind{Type: gosnmp.TimeTicks, Value: uint32(100)},
			"100 (1s)",
		},
		{
			"TimeTicks uint",
			trap.Varbind{Type: gosnmp.TimeTicks, Value: uint(6000)},
			"6000 (1m0s)",
		},
		{
			"TimeTicks wrong type falls through",
			trap.Varbind{Type: gosnmp.TimeTicks, Value: "bad"},
			"bad",
		},

		// OctetString
		{
			"OctetString printable",
			trap.Varbind{Type: gosnmp.OctetString, Value: []byte("test")},
			"test",
		},
		{
			"OctetString with null",
			trap.Varbind{Type: gosnmp.OctetString, Value: []byte{0x00}},
			"0x00",
		},
		{
			"OctetString wrong type falls through",
			trap.Varbind{Type: gosnmp.OctetString, Value: 42},
			"42",
		},

		// ObjectIdentifier (nil loader returns raw OID)
		{
			"ObjectIdentifier string",
			trap.Varbind{Type: gosnmp.ObjectIdentifier, Value: ".1.3.6.1.2"},
			".1.3.6.1.2",
		},
		{
			"ObjectIdentifier wrong type falls through",
			trap.Varbind{Type: gosnmp.ObjectIdentifier, Value: 42},
			"42",
		},

		// IPAddress
		{
			"IPAddress string",
			trap.Varbind{Type: gosnmp.IPAddress, Value: "192.168.1.1"},
			"192.168.1.1",
		},
		{
			"IPAddress wrong type falls through",
			trap.Varbind{Type: gosnmp.IPAddress, Value: 42},
			"42",
		},

		// Integer (nil loader — no enum lookup)
		{
			"Integer default",
			trap.Varbind{Type: gosnmp.Integer, Value: int(99)},
			"99",
		},

		// Counter32
		{
			"Counter32 default",
			trap.Varbind{Type: gosnmp.Counter32, Value: uint32(500)},
			"500",
		},

		// Null
		{
			"Null",
			trap.Varbind{Type: gosnmp.Null, Value: nil},
			"null",
		},

		// Boolean
		{
			"Boolean true",
			trap.Varbind{Type: gosnmp.Boolean, Value: true},
			"true",
		},
		{
			"Boolean false",
			trap.Varbind{Type: gosnmp.Boolean, Value: false},
			"false",
		},
		{
			"Boolean wrong type falls through",
			trap.Varbind{Type: gosnmp.Boolean, Value: int(1)},
			"1",
		},

		// Opaque
		{
			"Opaque bytes",
			trap.Varbind{Type: gosnmp.Opaque, Value: []byte{0xde, 0xad}},
			"0xdead",
		},
		{
			"Opaque wrong type falls through",
			trap.Varbind{Type: gosnmp.Opaque, Value: "raw"},
			"raw",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValue(nil, tt.vb)
			if got != tt.expected {
				t.Errorf("FormatValue(nil, %+v) = %q, want %q", tt.vb, got, tt.expected)
			}
		})
	}
}
