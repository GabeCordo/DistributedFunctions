// Package database
//
// Copyright (c) 2025-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  types.go
package database

import (
	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/pipeline"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic"
)

type UseCases struct {
	PipelineDatabase  pipeline.Database
	StatisticDatabase statistic.Database
	JobDatabase       job.Database
	Logger            logging.Logger
}
