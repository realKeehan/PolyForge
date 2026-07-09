package kumi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestInstallHostedPackResolvesRelativeURL is the regression test for the
// "download failed with status 404" bug: a hosted pack whose download URL is
// root-relative ("/packs/<id>.polypack") must resolve against the website base
// and install cleanly, streaming progress + log events as it goes.
func TestInstallHostedPackResolvesRelativeURL(t *testing.T) {
	// A valid pack, served from a stand-in "website".
	packPath := buildTestPack(t, map[string]string{"mods/a.jar": "pretend jar bytes"}, true)
	packBytes, err := os.ReadFile(packPath)
	if err != nil {
		t.Fatal(err)
	}

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write(packBytes)
	}))
	defer srv.Close()

	// Point the download base at the test server, as if it were polyforge.dev.
	prevBase := downloadGatewayBase
	downloadGatewayBase = srv.URL
	defer func() { downloadGatewayBase = prevBase }()

	// Capture streamed events so we can assert progress/log actually fired.
	var mu sync.Mutex
	var progress, logs int
	svc := NewService()
	svc.SetEmitter(func(_ string, data ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		for _, d := range data {
			ev, ok := d.(InstallEvent)
			if !ok {
				continue
			}
			switch ev.Kind {
			case "progress":
				progress++
			case "log":
				logs++
			}
		}
	})

	target := t.TempDir()
	result, err := svc.installHostedPack(ExecutionPayload{
		Path:  target,
		Extra: map[string]string{"packUrl": "/packs/test.polypack", "packName": "Test Pack"},
	})
	if err != nil {
		t.Fatalf("installHostedPack returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("install failed; messages: %+v", result.Messages)
	}

	// The relative URL must have been fetched as /packs/test.polypack (resolved
	// against the base), not rejected or 404'd.
	if gotPath != "/packs/test.polypack" {
		t.Errorf("server saw path %q, want /packs/test.polypack", gotPath)
	}

	// The pack's overrides landed in the target.
	if _, err := os.Stat(filepath.Join(target, "mods", "a.jar")); err != nil {
		t.Errorf("expected extracted file mods/a.jar: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if progress == 0 {
		t.Error("expected at least one progress event, got none")
	}
	if logs == 0 {
		t.Error("expected at least one log event, got none")
	}
}

// TestInstallHostedPackMissingURL surfaces a clear error instead of downloading.
func TestInstallHostedPackMissingURL(t *testing.T) {
	svc := NewService()
	result, err := svc.installHostedPack(ExecutionPayload{Path: t.TempDir()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when no packUrl is provided")
	}
}
