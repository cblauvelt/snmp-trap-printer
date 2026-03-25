package mib

import (
	"slices"
	"testing"
)

// testdataMIBs is the path to the bundled MIB files used by loader tests.
// Relative to the package directory (internal/mib/).
const testdataMIBs = "../../testdata/mibs"

func TestNewLoader_NoPaths(t *testing.T) {
	l := NewLoader(nil)
	if l == nil {
		t.Fatal("expected non-nil loader")
	}
}

func TestNewLoader_NoPaths_LoadedModulesEmpty(t *testing.T) {
	l := NewLoader(nil)
	// No MIB paths → no modules loaded; loader must still be usable.
	if l.LoadedModules() == nil {
		// nil is acceptable; length must be zero
	} else if len(l.LoadedModules()) != 0 {
		t.Errorf("expected 0 loaded modules with no paths, got %d: %v", len(l.LoadedModules()), l.LoadedModules())
	}
}

func TestNewLoader_WithTestdata_LoadsAllModules(t *testing.T) {
	l := NewLoader([]string{testdataMIBs})

	loaded := l.LoadedModules()
	if len(loaded) != len(standardMIBs) {
		t.Errorf("expected %d modules loaded, got %d: %v", len(standardMIBs), len(loaded), loaded)
	}
}

func TestNewLoader_WithTestdata_LoadedModulesContainExpected(t *testing.T) {
	l := NewLoader([]string{testdataMIBs})
	loaded := l.LoadedModules()

	for _, want := range []string{"SNMPv2-SMI", "SNMPv2-TC", "SNMPv2-MIB", "IF-MIB"} {
		if !slices.Contains(loaded, want) {
			t.Errorf("expected module %q in LoadedModules, got %v", want, loaded)
		}
	}
}

func TestTranslate_KnownOID(t *testing.T) {
	l := NewLoader([]string{testdataMIBs})

	// 1.3.6.1.2.1.1.1 = sysDescr in SNMPv2-MIB
	name, ok := l.Translate("1.3.6.1.2.1.1.1")
	if !ok {
		t.Fatal("expected Translate to find OID 1.3.6.1.2.1.1.1")
	}
	if name != "SNMPv2-MIB::sysDescr" {
		t.Errorf("Translate(1.3.6.1.2.1.1.1) = %q, want %q", name, "SNMPv2-MIB::sysDescr")
	}
}

func TestTranslate_LeadingDot(t *testing.T) {
	l := NewLoader([]string{testdataMIBs})

	// Leading dot must be stripped before lookup.
	name, ok := l.Translate(".1.3.6.1.2.1.1.1")
	if !ok {
		t.Fatal("expected Translate to find OID .1.3.6.1.2.1.1.1")
	}
	if name != "SNMPv2-MIB::sysDescr" {
		t.Errorf("Translate(.1.3.6.1.2.1.1.1) = %q, want %q", name, "SNMPv2-MIB::sysDescr")
	}
}

func TestTranslate_NoMIBsLoaded_FallsBackToWellKnown(t *testing.T) {
	l := NewLoader(nil)

	// With no MIBs loaded, gosmi resolves standard OIDs to built-in well-known
	// ancestor nodes rather than returning false.
	name, ok := l.Translate("1.3.6.1.2.1.1.1")
	if !ok {
		t.Fatal("expected Translate to resolve via built-in well-known nodes")
	}
	if name == "SNMPv2-MIB::sysDescr" {
		t.Errorf("expected unresolved name (no MIBs loaded), got fully qualified %q", name)
	}
}

func TestTranslate_ScalarInstance(t *testing.T) {
	l := NewLoader([]string{testdataMIBs})

	// sysDescr.0 — scalar instance OID; base node is 1.3.6.1.2.1.1.1
	name, ok := l.Translate("1.3.6.1.2.1.1.1.0")
	if !ok {
		t.Fatal("expected Translate to find OID 1.3.6.1.2.1.1.1.0")
	}
	if name != "SNMPv2-MIB::sysDescr.0" {
		t.Errorf("Translate(1.3.6.1.2.1.1.1.0) = %q, want %q", name, "SNMPv2-MIB::sysDescr.0")
	}
}

func TestTranslate_TabularInstance(t *testing.T) {
	l := NewLoader([]string{testdataMIBs})

	// ifDescr.1 — tabular instance OID; base node is 1.3.6.1.2.1.2.2.1.2
	name, ok := l.Translate("1.3.6.1.2.1.2.2.1.2.1")
	if !ok {
		t.Fatal("expected Translate to find OID 1.3.6.1.2.1.2.2.1.2.1")
	}
	// RFC1213-MIB defines ifDescr at the same OID; either qualifier is valid.
	if name != "IF-MIB::ifDescr.1" && name != "RFC1213-MIB::ifDescr.1" {
		t.Errorf("Translate(1.3.6.1.2.1.2.2.1.2.1) = %q, want IF-MIB::ifDescr.1 or RFC1213-MIB::ifDescr.1", name)
	}
}

func TestTranslate_UnknownOID(t *testing.T) {
	l := NewLoader([]string{testdataMIBs})

	// OID under root 99 — outside any standard MIB tree, no ancestor node exists.
	oid := "99.1.2.3"
	name, ok := l.Translate(oid)
	if ok {
		t.Errorf("expected Translate to return false for out-of-tree OID, got name=%q", name)
	}
	if name != oid {
		t.Errorf("Translate(%q) = %q, want original OID back", oid, name)
	}
}
