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
