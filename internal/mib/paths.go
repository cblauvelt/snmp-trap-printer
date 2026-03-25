package mib

import (
	"log/slog"
	"os"
	"runtime"
	"strings"
)

// DefaultMIBPaths returns OS-appropriate default MIB search directories.
// Only paths that exist on disk are returned.
func DefaultMIBPaths() []string {
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

// ResolveMIBPaths merges extra CLI-supplied paths (prepended, higher priority)
// with OS defaults, deduplicates, and returns only paths that exist on disk.
func ResolveMIBPaths(extra []string) []string {
	combined := append(existingPaths(extra), DefaultMIBPaths()...)
	return deduplicate(combined)
}

// existingPaths filters a list to only paths that exist on disk, logging each result.
func existingPaths(paths []string) []string {
	var out []string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			slog.Debug("MIB path found", "path", p)
			out = append(out, p)
		} else {
			slog.Debug("MIB path not found, skipping", "path", p)
		}
	}
	return out
}

// deduplicate removes duplicate paths, preserving order and first occurrence.
// On Windows comparisons are case-insensitive; elsewhere they are case-sensitive.
func deduplicate(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		key := p
		if runtime.GOOS == "windows" {
			key = strings.ToLower(p)
		}
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			out = append(out, p)
		}
	}
	return out
}
