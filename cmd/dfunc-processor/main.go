// Package main
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  main.go
package main

import (
	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/cmd/dfunc-processor/controllers"
)

func main() {

	cli := commandline.NewCommandLine()

	start := cli.AddCommand("start", controllers.ExampleController{})
	start.SetCategory("processor").SetDescription("Start a processor that connects to a running dfunc machine")

	cli.Run()
}
