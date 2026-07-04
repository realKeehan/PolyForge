package main

import (
	_ "embed"
	"strings"

	"polyforge/internal/kumi"
)

// The repo-root VERSION file is the single source of truth for the app
// version. It is embedded at compile time and injected into the kumi
// package here; the frontend receives the same value via a Vite define
// (see frontend/vite.config.ts). Bump VERSION, rebuild, done.
//
//go:embed VERSION
var embeddedVersion string

func init() {
	if v := strings.TrimSpace(embeddedVersion); v != "" {
		kumi.AppVersion = v
	}
}
