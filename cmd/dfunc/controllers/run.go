// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  run.go
package controllers

import (
	"fmt"
	"strconv"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"

	commandline "github.com/GabeCordo/Commandline"
)

type RunController struct {
}

func (controller RunController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Stop) {
		controller.stop(cli)
	} else {
		controller.create(cli)
	}

	return commandline.Terminate
}

func (controller RunController) create(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	namespace := cli.NextArg()
	if namespace == commandline.FinalArg {
		fmt.Println("Missing 'namespace' Parameter: dfunc run [namespace] [pipeline]")
		return commandline.Terminate
	}

	pipeline := cli.NextArg()
	if pipeline == commandline.FinalArg {
		fmt.Printf("Missing 'pipeline' Parameter: dfunc run %s [pipeline]\n", namespace)
		return commandline.Terminate
	}

	core := "http://localhost:8136"
	runId, err := api.RunPipelineOnCore(core, namespace, pipeline)
	if err != nil {
		fmt.Printf("Failed to run the pipeline: %s\n", err)
		return commandline.Terminate
	}

	// the controller must only output the id of the run that was succesfully created
	// note: outputting other information makes it difficult for automated tools to use the output
	fmt.Println(runId)

	return commandline.Terminate
}

func (controller RunController) stop(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	runIdStr := cli.NextArg()
	if runIdStr == commandline.FinalArg {
		fmt.Println("Missing 'id' Parameter: dfunc stop run [id]")
		return commandline.Terminate
	}

	runId, err := strconv.ParseUint(runIdStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid 'id' Parameter: must be a uint64 but was '%s'\n", runIdStr)
		return commandline.Terminate
	}

	core := "http://localhost:8136"
	err = api.StopRun(core, runId)
	if err != nil {
		fmt.Printf("Failed to stop run %d: %s\n", runId, err)
	} else {
		fmt.Printf("Succesfully stopped run %d\n", runId)
	}

	return commandline.Terminate
}
