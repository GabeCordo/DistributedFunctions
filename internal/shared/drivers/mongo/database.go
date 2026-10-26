// Package mongo
//
// Copyright (c) 2025 Gabriel Cordovado
// All rights reserved
//
// Source file:  database.go
package mongo

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	database *mongo.Database
}

func (database Database) Collection(name string) Collection {
	c := database.database.Collection(name)
	return Collection{c}
}
