package trap

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/cblauvelt/snmp-trap-printer/internal/config"
	"github.com/gosnmp/gosnmp"
)

// TrapHandler is called for each received trap with the raw SNMP packet.
type TrapHandler func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr)

// Listener wraps a gosnmp TrapListener.
type Listener struct {
	cfg     *config.Config
	handler TrapHandler
	tl      *gosnmp.TrapListener
	done    chan struct{}
}

// New constructs a Listener from the given config and handler.
func New(cfg *config.Config, handler TrapHandler) *Listener {
	return &Listener{
		cfg:     cfg,
		handler: handler,
		done:    make(chan struct{}),
	}
}

// Start begins receiving traps and blocks until Stop is called.
// It returns an error immediately if the listener fails to bind.
func (l *Listener) Start() error {
	params := gosnmp.Default
	if l.cfg.V3 != nil {
		params = &gosnmp.GoSNMP{
			Version:       gosnmp.Version3,
			SecurityModel: gosnmp.UserSecurityModel,
			MsgFlags:      l.cfg.V3.SecurityLevel(),
			SecurityParameters: &gosnmp.UsmSecurityParameters{
				UserName:                 l.cfg.V3.Username,
				AuthenticationProtocol:   l.cfg.V3.AuthProtocol,
				AuthenticationPassphrase: l.cfg.V3.AuthPassword,
				PrivacyProtocol:          l.cfg.V3.PrivProtocol,
				PrivacyPassphrase:        l.cfg.V3.PrivPassword,
			},
			Logger: gosnmp.NewLogger(slogAdapter{}),
		}
	}

	l.tl = gosnmp.NewTrapListener()
	l.tl.OnNewTrap = func(p *gosnmp.SnmpPacket, addr *net.UDPAddr) {
		slog.Debug("received trap", "src", addr)
		l.handler(p, addr)
	}
	l.tl.Params = params

	addr := fmt.Sprintf("%s:%d", l.cfg.Address, l.cfg.Port)
	slog.Info("listening for SNMP traps", "addr", fmt.Sprintf("udp://%s", addr))

	errCh := make(chan error, 1)
	go func() {
		errCh <- l.tl.Listen(addr)
	}()

	// Return a bind error if it occurs before Stop is called.
	// If Stop() was called, tl.Listen() will also return an error (closed
	// network connection); treat that as a clean exit.
	select {
	case err := <-errCh:
		select {
		case <-l.done:
			return nil
		default:
			return err
		}
	case <-l.done:
		return nil
	}
}

// Stop shuts down the listener cleanly.
func (l *Listener) Stop() {
	if l.tl != nil {
		l.tl.Close()
	}
	close(l.done)
}

// Dispatch parses a raw SNMP packet into a Trap struct.
func Dispatch(p *gosnmp.SnmpPacket, addr *net.UDPAddr) *Trap {
	switch p.Version {
	case gosnmp.Version1:
		return ParseV1(p, addr.String())
	case gosnmp.Version2c:
		return ParseV2c(p, addr.String())
	case gosnmp.Version3:
		return ParseV3(p, addr.String())
	default:
		slog.Warn("unknown SNMP version", "version", p.Version)
		return nil
	}
}
