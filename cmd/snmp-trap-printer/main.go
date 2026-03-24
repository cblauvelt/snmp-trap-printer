package main

import (
	"fmt"
	"log/slog"
	"os"

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

	handler := func(t *trap.Trap) {
		// Resolve OID names for all varbinds.
		for i := range t.Varbinds {
			if name, ok := loader.Translate(t.Varbinds[i].OID); ok {
				t.Varbinds[i].Name = name
			}
		}
		if err := formatter.Format(os.Stdout, t); err != nil {
			slog.Error("format error", "err", err)
		}
	}

	var params *gosnmp.GoSNMP
	if cfg.V3 != nil && cfg.V3.Username != "" {
		params = &gosnmp.GoSNMP{
			Version:       gosnmp.Version3,
			SecurityModel: gosnmp.UserSecurityModel,
			MsgFlags:      cfg.V3.SecurityLevel(),
			SecurityParameters: &gosnmp.UsmSecurityParameters{
				UserName:                 cfg.V3.Username,
				AuthenticationProtocol:   cfg.V3.AuthProtocol,
				AuthenticationPassphrase: cfg.V3.AuthPassword,
				PrivacyProtocol:          cfg.V3.PrivProtocol,
				PrivacyPassphrase:        cfg.V3.PrivPassword,
			},
		}
	}

	listener := trap.NewListener(cfg.Address, cfg.Port, params, handler)
	if err := listener.Listen(); err != nil {
		fmt.Fprintf(os.Stderr, "listener error: %v\n", err)
		os.Exit(1)
	}
}
