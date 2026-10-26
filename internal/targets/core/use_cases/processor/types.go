// Package processor
//
// Copyright (c) 2025-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  types.go
package processor

import (
	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/processor"
)

type UseCases struct {
	ProcessorTable *processor.Table
	Logger         logging.Logger
}
