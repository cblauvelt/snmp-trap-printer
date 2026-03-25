package output

import (
	"fmt"
	"io"

	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
	"github.com/gosnmp/gosnmp"
)

const divider = "--------------------------------------------------------------------------------"

// HumanFormatter writes labeled KV blocks to stdout, one per trap.
type HumanFormatter struct{}

// NewHumanFormatter returns a HumanFormatter.
func NewHumanFormatter() *HumanFormatter { return &HumanFormatter{} }

// Format writes a human-readable block for t.
func (h *HumanFormatter) Format(w io.Writer, t *trap.Trap) error {
	fmt.Fprintln(w, divider)
	fmt.Fprintf(w, "Version:    %s\n", versionString(t.Version))
	fmt.Fprintf(w, "Source:     %s\n", t.SourceIP)

	if t.Community != "" {
		fmt.Fprintf(w, "Community:  %s\n", t.Community)
	}
	if t.Enterprise != "" {
		fmt.Fprintf(w, "Enterprise: %s\n", t.Enterprise)
	}
	if t.AgentAddress != "" {
		fmt.Fprintf(w, "Agent:      %s\n", t.AgentAddress)
	}
	if t.Version == gosnmp.Version1 {
		fmt.Fprintf(w, "Generic:    %d\n", t.GenericType)
		fmt.Fprintf(w, "Specific:   %d\n", t.SpecificType)
	}
	if t.Timestamp > 0 {
		fmt.Fprintf(w, "Timestamp:  %d timeticks\n", t.Timestamp)
	}
	if t.OID != "" {
		fmt.Fprintf(w, "Trap OID:   %s\n", t.OID)
	}
	if t.SecurityName != "" {
		fmt.Fprintf(w, "SecName:    %s\n", t.SecurityName)
	}
	if t.ContextName != "" {
		fmt.Fprintf(w, "Context:    %s\n", t.ContextName)
	}

	if len(t.Varbinds) > 0 {
		fmt.Fprintln(w, "Varbinds:")
		for _, vb := range t.Varbinds {
			name := vb.OID
			if vb.Name != "" {
				name = vb.Name
			}
			val := vb.FormattedValue
			if val == "" {
				val = fmt.Sprintf("%v", vb.Value)
			}
			fmt.Fprintf(w, "  %-40s %s\n", name, val)
		}
	}
	fmt.Fprintln(w, divider)
	return nil
}

func versionString(v gosnmp.SnmpVersion) string {
	switch v {
	case gosnmp.Version1:
		return "v1"
	case gosnmp.Version2c:
		return "v2c"
	case gosnmp.Version3:
		return "v3"
	default:
		return fmt.Sprintf("unknown(%d)", v)
	}
}

// ensure interface is satisfied
var _ Formatter = (*HumanFormatter)(nil)
