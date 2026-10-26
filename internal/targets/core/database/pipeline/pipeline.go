// Package pipeline
//
// Copyright (c) 2024-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  pipeline.go
package pipeline

import (
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/ScalingFunctions"
)

type Pipeline struct {
	Namespace  string                       `json:"namespace"`
	Identifier string                       `json:"identifier"`
	Data       *ScalingFunctions.PipelineIR `json:"data"`
}

type Database interface {
	Get(filter database.Filter) []*ScalingFunctions.PipelineIR
	Create(filter database.Filter, record *ScalingFunctions.PipelineIR) (string, error)
	Replace(filter database.Filter, record *ScalingFunctions.PipelineIR) error
	Delete(filter database.Filter) error
	Distinct(filter database.Filter) ([]any, error)
	Save(path string) error
	Load(path string) error
	Print()
}
