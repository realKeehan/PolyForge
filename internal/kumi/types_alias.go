package kumi

import ktypes "polyforge/internal/kumi/types"

type (
	OptionDescriptor        = ktypes.OptionDescriptor
	ExecutionPayload        = ktypes.ExecutionPayload
	LogEntry                = ktypes.LogEntry
	ActionResult            = ktypes.ActionResult
	ModrinthCloneRequest    = ktypes.ModrinthCloneRequest
	ExecutableSearchRequest = ktypes.ExecutableSearchRequest
	ApplicationInfo         = ktypes.ApplicationInfo
	InstallMode             = ktypes.InstallMode
)

const (
	ModeAuto      = ktypes.ModeAuto
	ModeInstall   = ktypes.ModeInstall
	ModeUpdate    = ktypes.ModeUpdate
	ModeReinstall = ktypes.ModeReinstall
	ModeClean     = ktypes.ModeClean
)

var (
	NewResult = ktypes.NewResult
)
