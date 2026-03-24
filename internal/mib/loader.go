package mib

import (
	"log/slog"
	"strings"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)

// Loader holds MIB loading state. Must be created once at startup via NewLoader.
type Loader struct {
	paths  []string
	loaded int
}

// NewLoader initializes gosmi and loads MIBs from default + extra paths.
func NewLoader(extraPaths []string) *Loader {
	l := &Loader{
		paths: resolvePaths(extraPaths),
	}

	gosmi.Init()

	for _, p := range l.paths {
		gosmi.AppendPath(p)
	}

	loaded := gosmi.GetLoadedModules()
	l.loaded = len(loaded)
	slog.Info("MIB loader initialized", "paths", l.paths, "modules", l.loaded)

	return l
}

// Translate converts a dotted OID string to a human-readable MIB name.
// Returns the original OID if not found.
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
	return node.RenderQualified(), true
}
