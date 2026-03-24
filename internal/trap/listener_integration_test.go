//go:build integration

package trap_test

import (
	"net"
	"testing"
	"time"

	"github.com/cblauvelt/snmp-trap-printer/internal/config"
	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
	"github.com/gosnmp/gosnmp"
)

// freeUDPPort finds an available UDP port by briefly binding to :0,
// reading the assigned port, and closing. There is a small race window
// between close and the listener binding — acceptable in isolated test runs.
func freeUDPPort(t *testing.T) uint16 {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freeUDPPort: %v", err)
	}
	port := conn.(*net.UDPConn).LocalAddr().(*net.UDPAddr).Port
	conn.Close()
	return uint16(port)
}

func newTestListener(t *testing.T, port uint16, handler trap.TrapHandler) (*trap.Listener, chan error) {
	t.Helper()
	cfg := &config.Config{
		Address: "127.0.0.1",
		Port:    port,
		V3:      &config.V3Config{},
	}
	l := trap.New(cfg, handler)
	errCh := make(chan error, 1)
	go func() { errCh <- l.Start() }()
	// Give the listener time to bind before sending traps.
	time.Sleep(50 * time.Millisecond)
	return l, errCh
}

// TestListenerReceivesV1Trap is the full end-to-end test: start a listener,
// send a real SNMPv1 trap over UDP, and verify all parsed fields.
func TestListenerReceivesV1Trap(t *testing.T) {
	port := freeUDPPort(t)

	received := make(chan *gosnmp.SnmpPacket, 1)
	addrs := make(chan *net.UDPAddr, 1)

	l, errCh := newTestListener(t, port, func(p *gosnmp.SnmpPacket, a *net.UDPAddr) {
		received <- p
		addrs <- a
	})
	defer l.Stop()

	// Build sender.
	sender := &gosnmp.GoSNMP{
		Target:    "127.0.0.1",
		Port:      port,
		Version:   gosnmp.Version1,
		Community: "public",
		Timeout:   2 * time.Second,
	}
	if err := sender.Connect(); err != nil {
		t.Fatalf("sender.Connect: %v", err)
	}
	defer sender.Conn.Close()

	// Send a v1 trap.
	trapPDU := gosnmp.SnmpTrap{
		Enterprise:   ".1.3.6.1.4.1.9",
		AgentAddress: "127.0.0.1",
		GenericTrap:  6,
		SpecificTrap: 1,
		Timestamp:    12345,
		Variables: []gosnmp.SnmpPDU{
			{Name: ".1.3.6.1.4.1.9.1.0", Type: gosnmp.Integer, Value: 99},
		},
	}
	if _, err := sender.SendTrap(trapPDU); err != nil {
		t.Fatalf("SendTrap: %v", err)
	}

	// Wait for the trap with a timeout.
	var pkt *gosnmp.SnmpPacket
	var addr *net.UDPAddr
	select {
	case pkt = <-received:
		addr = <-addrs
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for trap")
	}

	// Parse via Dispatch and verify.
	parsed := trap.Dispatch(pkt, addr)
	if parsed == nil {
		t.Fatal("Dispatch returned nil")
	}

	if parsed.Version != gosnmp.Version1 {
		t.Errorf("Version = %v, want Version1", parsed.Version)
	}
	if parsed.Community != "public" {
		t.Errorf("Community = %q, want %q", parsed.Community, "public")
	}
	if parsed.Enterprise != ".1.3.6.1.4.1.9" {
		t.Errorf("Enterprise = %q, want %q", parsed.Enterprise, ".1.3.6.1.4.1.9")
	}
	if parsed.GenericType != 6 {
		t.Errorf("GenericType = %d, want 6", parsed.GenericType)
	}
	if parsed.SpecificType != 1 {
		t.Errorf("SpecificType = %d, want 1", parsed.SpecificType)
	}
	if parsed.Timestamp != 12345 {
		t.Errorf("Timestamp = %d, want 12345", parsed.Timestamp)
	}
	if len(parsed.Varbinds) != 1 {
		t.Fatalf("len(Varbinds) = %d, want 1", len(parsed.Varbinds))
	}
	if parsed.Varbinds[0].OID != ".1.3.6.1.4.1.9.1.0" {
		t.Errorf("Varbinds[0].OID = %q, want %q", parsed.Varbinds[0].OID, ".1.3.6.1.4.1.9.1.0")
	}
	if parsed.Varbinds[0].Value != 99 {
		t.Errorf("Varbinds[0].Value = %v, want 99", parsed.Varbinds[0].Value)
	}
	// defer l.Stop() handles cleanup; Start() return is verified in TestListenerStop_CleanShutdown.
	_ = errCh
}

// TestListenerReceivesV3NoAuthNoPrivTrap verifies that a v3 noAuthNoPriv trap
// is received and correctly parsed when no credentials are configured.
func TestListenerReceivesV3NoAuthNoPrivTrap(t *testing.T) {
	port := freeUDPPort(t)

	received := make(chan *gosnmp.SnmpPacket, 1)
	addrs := make(chan *net.UDPAddr, 1)

	l, errCh := newTestListener(t, port, func(p *gosnmp.SnmpPacket, a *net.UDPAddr) {
		received <- p
		addrs <- a
	})
	defer l.Stop()

	sender := &gosnmp.GoSNMP{
		Target:        "127.0.0.1",
		Port:          port,
		Version:       gosnmp.Version3,
		SecurityModel: gosnmp.UserSecurityModel,
		MsgFlags:      gosnmp.NoAuthNoPriv,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName:                 "testuser",
			AuthoritativeEngineID:    string([]byte{0x80, 0x00, 0x00, 0x00, 0x01}),
			AuthoritativeEngineBoots: 1,
			AuthoritativeEngineTime:  1,
		},
		Timeout: 2 * time.Second,
	}
	if err := sender.Connect(); err != nil {
		t.Fatalf("sender.Connect: %v", err)
	}
	defer sender.Conn.Close()

	trapPDU := gosnmp.SnmpTrap{
		Variables: []gosnmp.SnmpPDU{
			{Name: ".1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: uint32(500)},
			{Name: ".1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.ObjectIdentifier, Value: ".1.3.6.1.4.1.9.9.1"},
			{Name: ".1.3.6.1.4.1.9.1.0", Type: gosnmp.Integer, Value: 42},
		},
	}
	if _, err := sender.SendTrap(trapPDU); err != nil {
		t.Fatalf("SendTrap: %v", err)
	}

	var pkt *gosnmp.SnmpPacket
	var addr *net.UDPAddr
	select {
	case pkt = <-received:
		addr = <-addrs
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for v3 trap")
	}

	parsed := trap.Dispatch(pkt, addr)
	if parsed == nil {
		t.Fatal("Dispatch returned nil")
	}

	if parsed.Version != gosnmp.Version3 {
		t.Errorf("Version = %v, want Version3", parsed.Version)
	}
	if parsed.OID != ".1.3.6.1.4.1.9.9.1" {
		t.Errorf("OID = %q, want %q", parsed.OID, ".1.3.6.1.4.1.9.9.1")
	}
	if parsed.Timestamp != 500 {
		t.Errorf("Timestamp = %d, want 500", parsed.Timestamp)
	}
	if parsed.SecurityName != "testuser" {
		t.Errorf("SecurityName = %q, want %q", parsed.SecurityName, "testuser")
	}
	if len(parsed.Varbinds) != 3 {
		t.Errorf("len(Varbinds) = %d, want 3", len(parsed.Varbinds))
	}
	_ = errCh
}

// TestListenerStop_CleanShutdown verifies that Stop causes Start to return nil.
func TestListenerStop_CleanShutdown(t *testing.T) {
	port := freeUDPPort(t)
	l, errCh := newTestListener(t, port, func(_ *gosnmp.SnmpPacket, _ *net.UDPAddr) {})

	l.Stop()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Start() returned non-nil error on clean shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Start() did not return after Stop within timeout")
	}
}

// TestListenerBindConflict verifies that binding two listeners to the same
// port causes the second Start to return a non-nil error.
func TestListenerBindConflict(t *testing.T) {
	port := freeUDPPort(t)

	l1, _ := newTestListener(t, port, func(_ *gosnmp.SnmpPacket, _ *net.UDPAddr) {})
	defer l1.Stop()

	// Second listener on the same port — should fail to bind.
	cfg2 := &config.Config{
		Address: "127.0.0.1",
		Port:    port,
		V3:      &config.V3Config{},
	}
	l2 := trap.New(cfg2, func(_ *gosnmp.SnmpPacket, _ *net.UDPAddr) {})
	errCh2 := make(chan error, 1)
	go func() { errCh2 <- l2.Start() }()
	defer l2.Stop()

	select {
	case err := <-errCh2:
		if err == nil {
			t.Error("second listener on same port should return bind error, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Error("second listener did not return a bind error within timeout")
	}
}
