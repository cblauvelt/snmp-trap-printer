package mib

import (
	"fmt"
	"math"
	"testing"

	"github.com/gosnmp/gosnmp"
)

// --- formatTimeTicks ---

func TestFormatTimeTicks(t *testing.T) {
	tests := []struct {
		ticks    uint32
		expected string
	}{
		{0, "0 days, 00:00:00.00"},
		{100, "0 days, 00:00:01.00"},   // 1 second
		{99, "0 days, 00:00:00.99"},    // centiseconds only
		{6000, "0 days, 00:01:00.00"},  // 1 minute
		{6001, "0 days, 00:01:00.01"},  // 1 minute + 1 centisecond
		{360000, "0 days, 01:00:00.00"}, // 1 hour
		{8640000, "1 days, 00:00:00.00"}, // 1 day
		{8640000 + 360000 + 6000 + 100, "1 days, 01:01:01.00"}, // 1 day, 1 hour, 1 min, 1 sec
		{2*8640000 + 2*360000 + 2*6000 + 2*100 + 50, "2 days, 02:02:02.50"},
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
				// Should start with "0x" for hex output
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

// --- FormatValue ---

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		typ      gosnmp.Asn1BER
		value    interface{}
		expected string
	}{
		// TimeTicks
		{"TimeTicks zero", gosnmp.TimeTicks, uint32(0), "0 days, 00:00:00.00"},
		{"TimeTicks 1 second", gosnmp.TimeTicks, uint32(100), "0 days, 00:00:01.00"},
		{"TimeTicks uint", gosnmp.TimeTicks, uint(6000), "0 days, 00:01:00.00"},
		{"TimeTicks wrong type falls through", gosnmp.TimeTicks, "bad", "bad"},

		// OctetString
		{"OctetString printable", gosnmp.OctetString, []byte("test"), "test"},
		{"OctetString with null", gosnmp.OctetString, []byte{0x00}, "0x00"},
		{"OctetString wrong type falls through", gosnmp.OctetString, 42, "42"},

		// ObjectIdentifier
		{"ObjectIdentifier string", gosnmp.ObjectIdentifier, ".1.3.6.1.2", ".1.3.6.1.2"},
		{"ObjectIdentifier wrong type falls through", gosnmp.ObjectIdentifier, 42, "42"},

		// IPAddress
		{"IPAddress string", gosnmp.IPAddress, "192.168.1.1", "192.168.1.1"},
		{"IPAddress wrong type falls through", gosnmp.IPAddress, 42, "42"},

		// Default (Integer, Counter32, etc.)
		{"Integer default", gosnmp.Integer, int(99), "99"},
		{"Counter32 default", gosnmp.Counter32, uint32(500), "500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValue(tt.typ, tt.value)
			if got != tt.expected {
				t.Errorf("FormatValue(%v, %v) = %q, want %q", tt.typ, tt.value, got, tt.expected)
			}
		})
	}
}
