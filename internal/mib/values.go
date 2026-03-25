package mib

import (
	"fmt"
	"strings"
	"time"

	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
	"github.com/gosnmp/gosnmp"
	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)

// FormatValue returns a human-readable string for a varbind value.
// It uses the loader to resolve enum names, units, and OID names from MIBs.
// If loader is nil or any lookup fails, it falls back to fmt.Sprintf("%v", vb.Value).
func FormatValue(loader *Loader, vb trap.Varbind) string {
	switch vb.Type {
	case gosnmp.TimeTicks:
		if v, ok := toUint32(vb.Value); ok {
			return formatTimeTicks(v)
		}
	case gosnmp.OctetString:
		if b, ok := vb.Value.([]byte); ok {
			return printableOrHex(b)
		}
	case gosnmp.ObjectIdentifier:
		if s, ok := vb.Value.(string); ok {
			if loader != nil {
				if name, ok := loader.Translate(s); ok {
					return name
				}
			}
			return s
		}
	case gosnmp.IPAddress:
		if s, ok := vb.Value.(string); ok {
			return s
		}
	case gosnmp.Integer:
		if loader != nil {
			if node, ok := getNodeForOID(vb.OID); ok && node.Type != nil && node.Type.Enum != nil {
				if intVal, ok := toInt64(vb.Value); ok {
					return node.Type.Enum.Name(intVal)
				}
			}
		}
	case gosnmp.Gauge32:
		if loader != nil {
			if node, ok := getNodeForOID(vb.OID); ok && node.Type != nil && node.Type.Units != "" {
				return fmt.Sprintf("%v %s", vb.Value, node.Type.Units)
			}
		}
	case gosnmp.Null:
		return "null"
	case gosnmp.Boolean:
		if b, ok := vb.Value.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
	case gosnmp.Opaque:
		if b, ok := vb.Value.([]byte); ok {
			return fmt.Sprintf("0x%x", b)
		}
	}
	return fmt.Sprintf("%v", vb.Value)
}

// getNodeForOID looks up a gosmi node by dotted-decimal OID string.
func getNodeForOID(oid string) (gosmi.SmiNode, bool) {
	oid = strings.TrimPrefix(oid, ".")
	parsed, err := types.OidFromString(oid)
	if err != nil {
		return gosmi.SmiNode{}, false
	}
	node, err := gosmi.GetNodeByOID(parsed)
	if err != nil {
		return gosmi.SmiNode{}, false
	}
	return node, true
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

// toInt64 converts a numeric interface{} to int64 for enum lookup.
func toInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case int:
		return int64(t), true
	case int32:
		return int64(t), true
	case int64:
		return t, true
	case uint:
		return int64(t), true
	case uint32:
		return int64(t), true
	}
	return 0, false
}

// formatTimeTicks converts hundredths-of-a-second timeticks to a readable string.
// Format: "<raw_ticks> (<duration>)" e.g. "12345 (2m3.45s)".
func formatTimeTicks(ticks uint32) string {
	d := time.Duration(ticks) * 10 * time.Millisecond
	return fmt.Sprintf("%d (%s)", ticks, d)
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
