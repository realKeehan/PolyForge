package kumi

import "testing"

func TestComparePackMods(t *testing.T) {
	installed := []PackMod{
		{Name: "sodium", Version: "0.5.2", SHA256: "aaa"},
		{Name: "lithium", Version: "0.11.0", SHA256: "bbb"},
		{Name: "removed-mod", Version: "1.0.0"},
		{Name: "rebuilt", Version: "2.0.0", SHA256: "old"},
	}
	latest := []PackMod{
		{Name: "sodium", Version: "0.5.3", SHA256: "ccc"},   // version bump
		{Name: "lithium", Version: "0.11.0", SHA256: "bbb"}, // unchanged
		{Name: "new-mod", Version: "1.0.0"},                 // added
		{Name: "rebuilt", Version: "2.0.0", SHA256: "new"},  // same version, new hash
	}

	diff := ComparePackMods(installed, latest)

	if len(diff.Added) != 1 || diff.Added[0].Name != "new-mod" {
		t.Errorf("Added = %v, want [new-mod]", diff.Added)
	}
	if len(diff.Removed) != 1 || diff.Removed[0].Name != "removed-mod" {
		t.Errorf("Removed = %v, want [removed-mod]", diff.Removed)
	}
	if len(diff.Changed) != 2 {
		t.Fatalf("Changed = %v, want sodium + rebuilt", diff.Changed)
	}
	changedNames := map[string]bool{}
	for _, m := range diff.Changed {
		changedNames[m.Name] = true
	}
	if !changedNames["sodium"] || !changedNames["rebuilt"] {
		t.Errorf("Changed names = %v, want sodium and rebuilt", changedNames)
	}
	if !diff.HasChanges() {
		t.Error("HasChanges() = false, want true")
	}

	empty := ComparePackMods(latest, latest)
	if empty.HasChanges() {
		t.Errorf("identical lists should have no changes, got %+v", empty)
	}
}

func TestComparePackModsKeysByModID(t *testing.T) {
	installed := []PackMod{
		// Display name and filename both changed upstream; the mod id is the
		// stable identity, so this must NOT read as removed+added.
		{ModID: "sodium", Name: "sodium-fabric", File: "sodium-fabric-0.5.2.jar", Version: "0.5.2", SHA256: "aaa"},
		{ModID: "lithium", Name: "Lithium", Version: "0.11.0", SHA256: "bbb"},
	}
	latest := []PackMod{
		{ModID: "sodium", Name: "Sodium", File: "sodium-0.5.3+mc1.20.1.jar", Version: "0.5.3", SHA256: "ccc"},
		{ModID: "lithium", Name: "Lithium", Version: "0.11.0", SHA256: "bbb"},
	}

	diff := ComparePackMods(installed, latest)
	if len(diff.Added) != 0 || len(diff.Removed) != 0 {
		t.Errorf("id-matched mods must not diff as added/removed, got added=%v removed=%v", diff.Added, diff.Removed)
	}
	if len(diff.Changed) != 1 || diff.Changed[0].ModID != "sodium" {
		t.Errorf("Changed = %v, want [sodium]", diff.Changed)
	}

	// An id never collides with a name-keyed legacy entry of the same text.
	mixed := ComparePackMods(
		[]PackMod{{Name: "sodium", Version: "0.5.2"}},
		[]PackMod{{ModID: "sodium", Name: "Sodium", Version: "0.5.2"}},
	)
	if len(mixed.Added) != 1 || len(mixed.Removed) != 1 {
		t.Errorf("legacy name key must not match a new id key, got %+v", mixed)
	}
}

func TestParsePackManifest(t *testing.T) {
	good := []byte(`{"schemaVersion":1,"id":"turtel-smp","version":"1.0.0","name":"Turtel SMP","loader":{"type":"quilt","version":"0.22.0"},"mods":[{"file":"a-1.0.jar","name":"a","version":"1.0"}],"overrides":{"folders":["mods"],"fileCount":1,"totalBytes":10}}`)
	m, err := ParsePackManifest(good)
	if err != nil {
		t.Fatalf("ParsePackManifest: %v", err)
	}
	if m.ID != "turtel-smp" || m.Loader.Type != "quilt" || len(m.Mods) != 1 {
		t.Errorf("unexpected manifest: %+v", m)
	}

	if _, err := ParsePackManifest([]byte(`{"name":"no id"}`)); err == nil {
		t.Error("expected error for manifest without id/version")
	}
	if _, err := ParsePackManifest([]byte(`not json`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
}
