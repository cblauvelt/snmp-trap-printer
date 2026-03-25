package mib

import (
	"log/slog"
	"strings"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)

// standardMIBs lists the MIB modules attempted on every startup.
var standardMIBs = []string{
	"SNMPv2-SMI",
	"SNMPv2-TC",
	"SNMPv2-MIB",
	"RFC1213-MIB",
	"IF-MIB",
	"IP-MIB",
	"TCP-MIB",
	"UDP-MIB",
	"HOST-RESOURCES-MIB",
	"NOTIFICATION-LOG-MIB",
}

// Loader holds MIB loading state. Must be created once at startup via NewLoader.
type Loader struct {
	paths         []string
	loadedModules []string
}

// NewLoader initializes gosmi and loads MIBs from default + extra paths.
// If no MIB paths exist or no modules load, it returns a valid empty loader
// and the application degrades gracefully to raw OID display.
func NewLoader(extraPaths []string) *Loader {
	l := &Loader{
		paths: ResolveMIBPaths(extraPaths),
	}

	gosmi.Exit()
	gosmi.Init()

	for _, p := range l.paths {
		gosmi.AppendPath(p)
	}

	for _, name := range standardMIBs {
		if _, err := gosmi.LoadModule(name); err != nil {
			slog.Warn("MIB module not loaded", "module", name, "err", err)
		} else {
			slog.Info("MIB module loaded", "module", name)
			l.loadedModules = append(l.loadedModules, name)
		}
	}

	slog.Info("MIB loader initialized", "paths", l.paths, "modules", len(l.loadedModules))

	return l
}

// LoadedModules returns the names of successfully loaded MIB modules.
func (l *Loader) LoadedModules() []string {
	return l.loadedModules
}

// Translate converts a dotted OID string to a human-readable MIB name.
// Returns "MODULE::objectName.instance" for instance OIDs (e.g. sysDescr.0),
// or the original OID string if no MIB node is found.
func (l *Loader) Translate(oid string) (string, bool) {
	oid = strings.TrimPrefix(oid, ".")
	parsed, err := types.OidFromString(oid)
	if err != nil {
		return oid, false
	}
	node, err := gosmi.GetNodeByOID(parsed)
	if err != nil {
		return oid, false
	}
	name := node.RenderQualified()
	nodeOidStr := node.Oid.String()
	if oid != nodeOidStr && strings.HasPrefix(oid, nodeOidStr+".") {
		name += "." + oid[len(nodeOidStr)+1:]
	}
	return name, true
}
