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
)

var (
	NewResult = ktypes.NewResult
)
