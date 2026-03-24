package trap

import (
	"net"
	"testing"

	"github.com/gosnmp/gosnmp"
)

// --- ParseV1 ---

func TestParseV1_Basic(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version1,
		Community: "public",
		SnmpTrap: gosnmp.SnmpTrap{
			Enterprise:   ".1.3.6.1.4.1.9",
			AgentAddress: "10.0.0.1",
			GenericTrap:  6,
			SpecificTrap: 1,
			Timestamp:    12345,
		},
		Variables: []gosnmp.SnmpPDU{
			{Name: ".1.3.6.1.4.1.9.1.0", Type: gosnmp.Integer, Value: 42},
		},
	}

	trap := ParseV1(p, "192.168.1.5:54321")

	if trap.Version != gosnmp.Version1 {
		t.Errorf("Version = %v, want Version1", trap.Version)
	}
	if trap.SourceIP != "192.168.1.5:54321" {
		t.Errorf("SourceIP = %q, want %q", trap.SourceIP, "192.168.1.5:54321")
	}
	if trap.Community != "public" {
		t.Errorf("Community = %q, want %q", trap.Community, "public")
	}
	if trap.Enterprise != ".1.3.6.1.4.1.9" {
		t.Errorf("Enterprise = %q, want %q", trap.Enterprise, ".1.3.6.1.4.1.9")
	}
	if trap.AgentAddress != "10.0.0.1" {
		t.Errorf("AgentAddress = %q, want %q", trap.AgentAddress, "10.0.0.1")
	}
	if trap.GenericType != 6 {
		t.Errorf("GenericType = %d, want 6", trap.GenericType)
	}
	if trap.SpecificType != 1 {
		t.Errorf("SpecificType = %d, want 1", trap.SpecificType)
	}
	if trap.Timestamp != 12345 {
		t.Errorf("Timestamp = %d, want 12345", trap.Timestamp)
	}
	if trap.OID != "" {
		t.Errorf("OID = %q, want empty", trap.OID)
	}
	if trap.SecurityName != "" {
		t.Errorf("SecurityName = %q, want empty", trap.SecurityName)
	}
	if len(trap.Varbinds) != 1 {
		t.Fatalf("len(Varbinds) = %d, want 1", len(trap.Varbinds))
	}
	if trap.Varbinds[0].OID != ".1.3.6.1.4.1.9.1.0" {
		t.Errorf("Varbinds[0].OID = %q, want %q", trap.Varbinds[0].OID, ".1.3.6.1.4.1.9.1.0")
	}
}

func TestParseV1_NoVarbinds(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version1,
		Variables: []gosnmp.SnmpPDU{},
	}
	trap := ParseV1(p, "10.0.0.1:162")
	if trap.Varbinds == nil {
		t.Error("Varbinds should be non-nil for empty Variables")
	}
	if len(trap.Varbinds) != 0 {
		t.Errorf("len(Varbinds) = %d, want 0", len(trap.Varbinds))
	}
}

// --- ParseV2c ---

func TestParseV2c_Basic(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "private",
		Variables: []gosnmp.SnmpPDU{
			{Name: ".1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: uint32(360000)},
			{Name: ".1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.ObjectIdentifier, Value: ".1.3.6.1.4.1.9.9.1"},
			{Name: ".1.3.6.1.4.1.9.1.0", Type: gosnmp.Integer, Value: 42},
		},
	}

	trap := ParseV2c(p, "10.0.0.2:162")

	if trap.Version != gosnmp.Version2c {
		t.Errorf("Version = %v, want Version2c", trap.Version)
	}
	if trap.Community != "private" {
		t.Errorf("Community = %q, want %q", trap.Community, "private")
	}
	if trap.SourceIP != "10.0.0.2:162" {
		t.Errorf("SourceIP = %q, want %q", trap.SourceIP, "10.0.0.2:162")
	}
	if trap.Timestamp != 360000 {
		t.Errorf("Timestamp = %d, want 360000", trap.Timestamp)
	}
	if trap.OID != ".1.3.6.1.4.1.9.9.1" {
		t.Errorf("OID = %q, want %q", trap.OID, ".1.3.6.1.4.1.9.9.1")
	}
	// extractV2cFields reads but does not remove vars — all 3 are in Varbinds
	if len(trap.Varbinds) != 3 {
		t.Errorf("len(Varbinds) = %d, want 3", len(trap.Varbinds))
	}
	if trap.Enterprise != "" {
		t.Errorf("Enterprise = %q, want empty", trap.Enterprise)
	}
	if trap.SecurityName != "" {
		t.Errorf("SecurityName = %q, want empty", trap.SecurityName)
	}
}

func TestParseV2c_MissingSysUpTime(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version: gosnmp.Version2c,
		Variables: []gosnmp.SnmpPDU{
			{Name: ".1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.ObjectIdentifier, Value: ".1.3.6.1.4.1.9.9.1"},
		},
	}
	trap := ParseV2c(p, "10.0.0.1:162")
	if trap.Timestamp != 0 {
		t.Errorf("Timestamp = %d, want 0", trap.Timestamp)
	}
	if trap.OID != ".1.3.6.1.4.1.9.9.1" {
		t.Errorf("OID = %q, want %q", trap.OID, ".1.3.6.1.4.1.9.9.1")
	}
}

func TestParseV2c_MissingTrapOID(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version: gosnmp.Version2c,
		Variables: []gosnmp.SnmpPDU{
			{Name: ".1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: uint32(1000)},
		},
	}
	trap := ParseV2c(p, "10.0.0.1:162")
	if trap.OID != "" {
		t.Errorf("OID = %q, want empty", trap.OID)
	}
	if trap.Timestamp != 1000 {
		t.Errorf("Timestamp = %d, want 1000", trap.Timestamp)
	}
}

func TestParseV2c_EmptyVarbinds(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Variables: []gosnmp.SnmpPDU{},
	}
	trap := ParseV2c(p, "10.0.0.1:162")
	if trap.Timestamp != 0 {
		t.Errorf("Timestamp = %d, want 0", trap.Timestamp)
	}
	if trap.OID != "" {
		t.Errorf("OID = %q, want empty", trap.OID)
	}
	if len(trap.Varbinds) != 0 {
		t.Errorf("len(Varbinds) = %d, want 0", len(trap.Varbinds))
	}
}

// --- ParseV3 ---

func TestParseV3_Basic(t *testing.T) {
	// ContextEngineID is a string in gosnmp.SnmpPacket; fmt.Sprintf("%x", s)
	// hex-encodes each byte, so "\x80\x00\x1f\x88\x04" → "80001f8804".
	p := &gosnmp.SnmpPacket{
		Version: gosnmp.Version3,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName: "trapuser",
		},
		ContextName:     "mycontext",
		ContextEngineID: "\x80\x00\x1f\x88\x04",
		Variables: []gosnmp.SnmpPDU{
			{Name: ".1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: uint32(500)},
			{Name: ".1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.ObjectIdentifier, Value: ".1.3.6.1.4.1.99"},
		},
	}

	trap := ParseV3(p, "172.16.0.1:162")

	if trap.Version != gosnmp.Version3 {
		t.Errorf("Version = %v, want Version3", trap.Version)
	}
	if trap.SecurityName != "trapuser" {
		t.Errorf("SecurityName = %q, want %q", trap.SecurityName, "trapuser")
	}
	if trap.ContextName != "mycontext" {
		t.Errorf("ContextName = %q, want %q", trap.ContextName, "mycontext")
	}
	if trap.ContextEngineID != "80001f8804" {
		t.Errorf("ContextEngineID = %q, want %q", trap.ContextEngineID, "80001f8804")
	}
	if trap.Timestamp != 500 {
		t.Errorf("Timestamp = %d, want 500", trap.Timestamp)
	}
	if trap.OID != ".1.3.6.1.4.1.99" {
		t.Errorf("OID = %q, want %q", trap.OID, ".1.3.6.1.4.1.99")
	}
	if trap.Community != "" {
		t.Errorf("Community = %q, want empty", trap.Community)
	}
}

func TestParseV3_EmptyContextEngineID(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version: gosnmp.Version3,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName: "user",
		},
		ContextEngineID: "",
	}
	trap := ParseV3(p, "10.0.0.1:162")
	if trap.ContextEngineID != "" {
		t.Errorf("ContextEngineID = %q, want empty string for empty engine ID", trap.ContextEngineID)
	}
}

// --- extractV2cFields ---

func TestExtractV2cFields_SysUpTimeWrongType(t *testing.T) {
	trap := &Trap{}
	vars := []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: "not-a-uint32"},
	}
	extractV2cFields(trap, vars)
	if trap.Timestamp != 0 {
		t.Errorf("Timestamp = %d, want 0 when value type is wrong", trap.Timestamp)
	}
}

func TestExtractV2cFields_OIDWrongType(t *testing.T) {
	trap := &Trap{}
	vars := []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.ObjectIdentifier, Value: 12345},
	}
	extractV2cFields(trap, vars)
	if trap.OID != "" {
		t.Errorf("OID = %q, want empty when value type is wrong", trap.OID)
	}
}

// --- convertVarbinds ---

func TestConvertVarbinds_AllTypes(t *testing.T) {
	vars := []gosnmp.SnmpPDU{
		{Name: ".1.1", Type: gosnmp.Integer, Value: int(10)},
		{Name: ".1.2", Type: gosnmp.OctetString, Value: []byte("hello")},
		{Name: ".1.3", Type: gosnmp.ObjectIdentifier, Value: ".1.3.6"},
		{Name: ".1.4", Type: gosnmp.IPAddress, Value: "192.168.1.1"},
		{Name: ".1.5", Type: gosnmp.TimeTicks, Value: uint32(100)},
	}
	vbs := convertVarbinds(vars)
	if len(vbs) != 5 {
		t.Fatalf("len(vbs) = %d, want 5", len(vbs))
	}
	for i, vb := range vbs {
		if vb.OID != vars[i].Name {
			t.Errorf("[%d] OID = %q, want %q", i, vb.OID, vars[i].Name)
		}
		if vb.Type != vars[i].Type {
			t.Errorf("[%d] Type = %v, want %v", i, vb.Type, vars[i].Type)
		}
		if vb.Name != "" {
			t.Errorf("[%d] Name = %q, want empty (MIB resolution not done by parser)", i, vb.Name)
		}
	}
}

func TestConvertVarbinds_Empty(t *testing.T) {
	vbs := convertVarbinds([]gosnmp.SnmpPDU{})
	if vbs == nil {
		t.Error("convertVarbinds should return non-nil slice for empty input")
	}
	if len(vbs) != 0 {
		t.Errorf("len(vbs) = %d, want 0", len(vbs))
	}
}

// --- Dispatch ---

func mustUDPAddr(t *testing.T, addr string) *net.UDPAddr {
	t.Helper()
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("ResolveUDPAddr(%q): %v", addr, err)
	}
	return a
}

func TestDispatch_V1(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version1,
		Community: "public",
	}
	trap := Dispatch(p, mustUDPAddr(t, "10.0.0.1:162"))
	if trap == nil {
		t.Fatal("Dispatch returned nil for Version1")
	}
	if trap.Version != gosnmp.Version1 {
		t.Errorf("Version = %v, want Version1", trap.Version)
	}
}

func TestDispatch_V2c(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "public",
	}
	trap := Dispatch(p, mustUDPAddr(t, "10.0.0.1:162"))
	if trap == nil {
		t.Fatal("Dispatch returned nil for Version2c")
	}
	if trap.Version != gosnmp.Version2c {
		t.Errorf("Version = %v, want Version2c", trap.Version)
	}
}

func TestDispatch_V3(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version: gosnmp.Version3,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName: "user",
		},
	}
	trap := Dispatch(p, mustUDPAddr(t, "10.0.0.1:162"))
	if trap == nil {
		t.Fatal("Dispatch returned nil for Version3")
	}
	if trap.Version != gosnmp.Version3 {
		t.Errorf("Version = %v, want Version3", trap.Version)
	}
}

func TestDispatch_UnknownVersion(t *testing.T) {
	p := &gosnmp.SnmpPacket{
		Version: gosnmp.SnmpVersion(99),
	}
	trap := Dispatch(p, mustUDPAddr(t, "10.0.0.1:162"))
	if trap != nil {
		t.Errorf("Dispatch returned non-nil for unknown version: %+v", trap)
	}
}
