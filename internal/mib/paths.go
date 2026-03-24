package mib

import (
	"log/slog"
	"os"
	"runtime"
)

// defaultPaths returns OS-aware default MIB search directories.
func defaultPaths() []string {
	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{
			"/usr/share/snmp/mibs",
			"/usr/local/share/snmp/mibs",
		}
	case "darwin":
		candidates = []string{
			"/usr/local/share/snmp/mibs",
			"/opt/homebrew/share/snmp/mibs",
		}
	case "windows":
		candidates = []string{
			`C:\usr\mibs`,
		}
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			candidates = append(candidates, appdata+`\snmp\mibs`)
		}
	}
	return existingPaths(candidates)
}

// existingPaths filters a list to only paths that exist on disk.
func existingPaths(paths []string) []string {
	var out []string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			out = append(out, p)
		} else {
			slog.Debug("MIB path not found, skipping", "path", p)
		}
	}
	return out
}

// resolvePaths merges default OS paths with any user-supplied extra paths,
// returning only those that exist on disk.
func resolvePaths(extra []string) []string {
	return append(defaultPaths(), existingPaths(extra)...)
}
