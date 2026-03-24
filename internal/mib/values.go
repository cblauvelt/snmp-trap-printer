package mib

import (
	"fmt"

	"github.com/gosnmp/gosnmp"
)

// FormatValue returns a human-readable string for a varbind value.
func FormatValue(typ gosnmp.Asn1BER, value interface{}) string {
	switch typ {
	case gosnmp.TimeTicks:
		if v, ok := toUint32(value); ok {
			return formatTimeTicks(v)
		}
	case gosnmp.OctetString:
		if b, ok := value.([]byte); ok {
			return printableOrHex(b)
		}
	case gosnmp.ObjectIdentifier:
		if s, ok := value.(string); ok {
			return s
		}
	case gosnmp.IPAddress:
		if s, ok := value.(string); ok {
			return s
		}
	}
	return fmt.Sprintf("%v", value)
}

func toUint32(v interface{}) (uint32, bool) {
	switch t := v.(type) {
	case uint32:
		return t, true
	case uint:
		return uint32(t), true
	case int:
		return uint32(t), true
	}
	return 0, false
}

// formatTimeTicks converts hundredths-of-a-second timeticks to a readable string.
func formatTimeTicks(ticks uint32) string {
	total := uint(ticks)
	days := total / 8640000
	total %= 8640000
	hours := total / 360000
	total %= 360000
	minutes := total / 6000
	total %= 6000
	seconds := total / 100
	centis := total % 100
	return fmt.Sprintf("%d days, %02d:%02d:%02d.%02d", days, hours, minutes, seconds, centis)
}

// printableOrHex returns the string as-is if all bytes are printable ASCII,
// otherwise returns a hex dump.
func printableOrHex(b []byte) string {
	for _, c := range b {
		if c < 0x20 || c > 0x7e {
			return fmt.Sprintf("0x%x", b)
		}
	}
	return string(b)
}
