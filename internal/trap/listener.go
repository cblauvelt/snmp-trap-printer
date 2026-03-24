package trap

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/gosnmp/gosnmp"
)

// Handler is called for each received trap.
type Handler func(t *Trap)

// Listener wraps a gosnmp TrapListener.
type Listener struct {
	address string
	port    uint16
	params  *gosnmp.GoSNMP
	handler Handler
}

// NewListener constructs a Listener. params may be nil for v1/v2c-only setups.
func NewListener(address string, port uint16, params *gosnmp.GoSNMP, handler Handler) *Listener {
	return &Listener{
		address: address,
		port:    port,
		params:  params,
		handler: handler,
	}
}

// Listen starts receiving traps and blocks until an error occurs.
func (l *Listener) Listen() error {
	tl := gosnmp.NewTrapListener()
	tl.OnNewTrap = l.dispatch
	tl.Params = l.params
	if tl.Params == nil {
		tl.Params = gosnmp.Default
	}

	addr := fmt.Sprintf("%s:%d", l.address, l.port)
	slog.Info("listening for SNMP traps", "addr", addr)
	return tl.Listen(addr)
}

func (l *Listener) dispatch(p *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	var t *Trap
	switch p.Version {
	case gosnmp.Version1:
		t = ParseV1(p, addr.String())
	case gosnmp.Version2c:
		t = ParseV2c(p, addr.String())
	case gosnmp.Version3:
		t = ParseV3(p, addr.String())
	default:
		slog.Warn("unknown SNMP version", "version", p.Version)
		return
	}
	l.handler(t)
}
