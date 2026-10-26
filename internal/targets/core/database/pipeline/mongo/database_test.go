// Package mongo
//
// Copyright (c) 2025-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  database_test.go
package mongo

import (
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/drivers/mongo"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/pipeline"
)

func Test_StatisticMongoDatabase_Is_Implemented(t *testing.T) {

	m := mongo.NewDriver("")
	d, err := NewMongoDatabase(m)
	if err != nil {
		t.Fatal(err)
	}

	var i pipeline.Database
	i = d

	i.Print()
}
