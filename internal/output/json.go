package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
	"github.com/gosnmp/gosnmp"
)

// JSONFormatter writes one JSON object per trap (NDJSON).
type JSONFormatter struct{}

// NewJSONFormatter returns a JSONFormatter.
func NewJSONFormatter() *JSONFormatter { return &JSONFormatter{} }

type jsonTrap struct {
	Version         string      `json:"version"`
	SourceIP        string      `json:"source_ip"`
	Community       string      `json:"community,omitempty"`
	Enterprise      string      `json:"enterprise,omitempty"`
	AgentAddress    string      `json:"agent_address,omitempty"`
	GenericType     int         `json:"generic_type,omitempty"`
	SpecificType    int         `json:"specific_type,omitempty"`
	Timestamp       uint        `json:"timestamp,omitempty"`
	OID             string      `json:"oid,omitempty"`
	SecurityName    string      `json:"security_name,omitempty"`
	ContextName     string      `json:"context_name,omitempty"`
	ContextEngineID string      `json:"context_engine_id,omitempty"`
	Varbinds        []jsonVarbind `json:"varbinds,omitempty"`
}

type jsonVarbind struct {
	OID   string `json:"oid"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Format encodes t as a single JSON line.
func (j *JSONFormatter) Format(w io.Writer, t *trap.Trap) error {
	jt := jsonTrap{
		Version:         versionString(t.Version),
		SourceIP:        t.SourceIP,
		Community:       t.Community,
		Enterprise:      t.Enterprise,
		AgentAddress:    t.AgentAddress,
		GenericType:     t.GenericType,
		SpecificType:    t.SpecificType,
		Timestamp:       t.Timestamp,
		OID:             t.OID,
		SecurityName:    t.SecurityName,
		ContextName:     t.ContextName,
		ContextEngineID: t.ContextEngineID,
	}
	for _, vb := range t.Varbinds {
		val := vb.FormattedValue
		if val == "" {
			val = fmt.Sprintf("%v", vb.Value)
		}
		jt.Varbinds = append(jt.Varbinds, jsonVarbind{
			OID:   vb.OID,
			Name:  vb.Name,
			Type:  berTypeName(vb.Type),
			Value: val,
		})
	}
	enc := json.NewEncoder(w)
	return enc.Encode(jt)
}

func berTypeName(t gosnmp.Asn1BER) string {
	switch t {
	case gosnmp.Integer:
		return "Integer"
	case gosnmp.OctetString:
		return "OctetString"
	case gosnmp.ObjectIdentifier:
		return "ObjectIdentifier"
	case gosnmp.IPAddress:
		return "IPAddress"
	case gosnmp.Counter32:
		return "Counter32"
	case gosnmp.Gauge32:
		return "Gauge32"
	case gosnmp.TimeTicks:
		return "TimeTicks"
	case gosnmp.Counter64:
		return "Counter64"
	default:
		return "Unknown"
	}
}

var _ Formatter = (*JSONFormatter)(nil)
