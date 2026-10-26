// Package scheduler
//
// Copyright (c) 2025-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  types.go
package scheduler

import (
	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/scheduler/job"
	jobDatabase "github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

type UseCases struct {
	Scheduler    *job.Scheduler
	JobsDatabase jobDatabase.Database
	Logger       logging.Logger
}
