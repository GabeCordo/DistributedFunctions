// Package in_memory
//
// Copyright (c) 2025-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  database_test.go
package in_memory

import (
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalDatabase()

	var i statistic.Database
	i = d

	i.Print()
}
