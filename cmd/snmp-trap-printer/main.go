package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/cblauvelt/snmp-trap-printer/internal/config"
	"github.com/cblauvelt/snmp-trap-printer/internal/mib"
	"github.com/cblauvelt/snmp-trap-printer/internal/output"
	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
	"github.com/gosnmp/gosnmp"
)

func main() {
	cfg := config.Parse()

	loader := mib.NewLoader(cfg.MIBPaths)

	var formatter output.Formatter
	switch cfg.Output {
	case config.OutputJSON:
		formatter = output.NewJSONFormatter()
	default:
		formatter = output.NewHumanFormatter()
	}

	handler := func(p *gosnmp.SnmpPacket, addr *net.UDPAddr) {
		t := trap.Dispatch(p, addr)
		if t == nil {
			return
		}
		for i := range t.Varbinds {
			if name, ok := loader.Translate(t.Varbinds[i].OID); ok {
				t.Varbinds[i].Name = name
			}
			t.Varbinds[i].FormattedValue = mib.FormatValue(loader, t.Varbinds[i])
		}
		if err := formatter.Format(os.Stdout, t); err != nil {
			slog.Error("format error", "err", err)
		}
	}

	listener := trap.New(cfg, handler)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		listener.Stop()
	}()

	if err := listener.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "listener error: %v\n", err)
		os.Exit(1)
	}
}
