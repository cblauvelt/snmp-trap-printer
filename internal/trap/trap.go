package trap

import "github.com/gosnmp/gosnmp"

// Varbind is a single variable binding from an SNMP trap PDU.
type Varbind struct {
	OID            string
	Name           string // resolved name; empty if OID not found in MIB
	FormattedValue string // human-readable value string; empty until enriched
	Type           gosnmp.Asn1BER
	Value          interface{}
}

// Trap is the version-neutral representation of a received SNMP trap.
// Version-specific fields are left zero/empty when not applicable.
type Trap struct {
	Version         gosnmp.SnmpVersion
	SourceIP        string
	Community       string // v1/v2c
	Enterprise      string // v1 only
	AgentAddress    string // v1 only
	GenericType     int    // v1 only
	SpecificType    int    // v1 only
	Timestamp       uint   // sysUpTime timeticks
	OID             string // v2c/v3 trap OID
	SecurityName    string // v3 only
	ContextName     string // v3 only
	ContextEngineID string // v3 only
	Varbinds        []Varbind
}
