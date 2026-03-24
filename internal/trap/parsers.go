package trap

import (
	"fmt"

	"github.com/gosnmp/gosnmp"
)

// ParseV1 converts a gosnmp v1 trap packet to a Trap.
func ParseV1(p *gosnmp.SnmpPacket, srcAddr string) *Trap {
	t := &Trap{
		Version:      gosnmp.Version1,
		SourceIP:     srcAddr,
		Community:    p.Community,
		Enterprise:   p.Enterprise,
		AgentAddress: p.AgentAddress,
		GenericType:  p.GenericTrap,
		SpecificType: p.SpecificTrap,
		Timestamp:    uint(p.Timestamp),
	}
	t.Varbinds = convertVarbinds(p.Variables)
	return t
}

// ParseV2c converts a gosnmp v2c trap packet to a Trap.
func ParseV2c(p *gosnmp.SnmpPacket, srcAddr string) *Trap {
	t := &Trap{
		Version:   gosnmp.Version2c,
		SourceIP:  srcAddr,
		Community: p.Community,
	}
	t.Varbinds = convertVarbinds(p.Variables)
	extractV2cFields(t, p.Variables)
	return t
}

// ParseV3 converts a gosnmp v3 trap packet to a Trap.
func ParseV3(p *gosnmp.SnmpPacket, srcAddr string) *Trap {
	t := &Trap{
		Version:         gosnmp.Version3,
		SourceIP:        srcAddr,
		SecurityName:    p.SecurityParameters.(*gosnmp.UsmSecurityParameters).UserName,
		ContextName:     p.ContextName,
		ContextEngineID: fmt.Sprintf("%x", p.ContextEngineID),
	}
	t.Varbinds = convertVarbinds(p.Variables)
	extractV2cFields(t, p.Variables) // v3 uses same varbind layout as v2c
	return t
}

// extractV2cFields pulls sysUpTime and snmpTrapOID from the standard v2c/v3 varbinds.
func extractV2cFields(t *Trap, vars []gosnmp.SnmpPDU) {
	for _, v := range vars {
		switch v.Name {
		case ".1.3.6.1.2.1.1.3.0": // sysUpTime
			if val, ok := v.Value.(uint32); ok {
				t.Timestamp = uint(val)
			}
		case ".1.3.6.1.6.3.1.1.4.1.0": // snmpTrapOID
			if val, ok := v.Value.(string); ok {
				t.OID = val
			}
		}
	}
}

func convertVarbinds(vars []gosnmp.SnmpPDU) []Varbind {
	vbs := make([]Varbind, len(vars))
	for i, v := range vars {
		vbs[i] = Varbind{
			OID:   v.Name,
			Type:  v.Type,
			Value: v.Value,
		}
	}
	return vbs
}
