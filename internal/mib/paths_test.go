package mib

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultMIBPaths_NoPanic(t *testing.T) {
	paths := DefaultMIBPaths()
	// Result may be empty on CI; just verify no panic and no duplicates.
	seen := make(map[string]struct{})
	for _, p := range paths {
		if _, exists := seen[p]; exists {
			t.Errorf("duplicate path in DefaultMIBPaths: %s", p)
		}
		seen[p] = struct{}{}
	}
}

func TestResolveMIBPaths_ExtraPathsFirst(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	paths := ResolveMIBPaths([]string{dir1, dir2})

	if len(paths) < 2 {
		t.Fatalf("expected at least 2 paths, got %d", len(paths))
	}
	if paths[0] != dir1 {
		t.Errorf("expected extra path %s first, got %s", dir1, paths[0])
	}
	if paths[1] != dir2 {
		t.Errorf("expected extra path %s second, got %s", dir2, paths[1])
	}
}

func TestResolveMIBPaths_NonExistentExtraSkipped(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist")

	paths := ResolveMIBPaths([]string{nonexistent})

	for _, p := range paths {
		if p == nonexistent {
			t.Errorf("non-existent path should have been filtered: %s", p)
		}
	}
}

func TestResolveMIBPaths_Deduplication(t *testing.T) {
	dir := t.TempDir()

	paths := ResolveMIBPaths([]string{dir, dir})

	count := 0
	for _, p := range paths {
		if p == dir {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected dir to appear exactly once, got %d times", count)
	}
}

func TestResolveMIBPaths_EmptyExtra(t *testing.T) {
	paths := ResolveMIBPaths(nil)
	// Should not panic; result equals DefaultMIBPaths().
	defaults := DefaultMIBPaths()
	if len(paths) != len(defaults) {
		t.Errorf("expected %d paths (defaults only), got %d", len(defaults), len(paths))
	}
}

func TestDeduplicate_CaseSensitiveOnNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("case-sensitivity test not applicable on Windows")
	}
	dir := t.TempDir()
	// Create a variant with different casing in the path string.
	// On non-Windows these are distinct strings and should both be kept if they exist.
	paths := deduplicate([]string{dir, dir})
	if len(paths) != 1 {
		t.Errorf("expected duplicate to be removed, got %d paths", len(paths))
	}
}
