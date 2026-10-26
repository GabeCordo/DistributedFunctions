// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  pipeline.go
package controllers

import (
	"fmt"
	"os"
	"os/exec"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/cmd/dfunc/local"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
	"github.com/GabeCordo/ScalingFunctions"
	"gopkg.in/yaml.v3"
)

type PipelineController struct {
}

func (controller PipelineController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	// TODO: Need to implement the pipeline command.
	return commandline.Terminate
}

func (controller PipelinesController) addPipelines(pipelines []*ScalingFunctions.PipelineIR) {

	core := local.GetCore()
	namespace := local.GetNamespace()

	// register each pipeline with the gateway
	for i, p := range pipelines {

		fmt.Printf("[%d] adding pipleine %s ", i, p.Identifier)
		onCore, err := api.IsPipelineOnCore(core, namespace, p.Identifier)
		if err != nil {
			fmt.Println("...skip")
			continue
		}

		if !onCore {
			if err = api.CreatePipelineOnCore(core, namespace, p); err != nil {
				fmt.Println("... failed")
			} else {
				fmt.Println("... done")
			}
		} else {
			fmt.Println("... already exists")
		}

	}
}

func (controller PipelinesController) deletePipelines(pipelines []*ScalingFunctions.PipelineIR) {

	core := local.GetCore()
	namespace := local.GetNamespace()

	for i, p := range pipelines {
		fmt.Printf("[%d] removing pipeline %s...", i, p.Identifier)
		if err := api.RemovePipelineFromCore(core, namespace, p.Identifier); err == nil {
			fmt.Println("... done")
		} else {
			fmt.Println("... skip")
		}
	}
}

func (controller PipelinesController) showPipeline(namespace, identifier string) {

	pipeline, err := api.GetPipeline(namespace, identifier)
	if err != nil {
		fmt.Printf("Failed to retrieve the pipeline '%s' from the dfunc machine: %s\n", identifier, err)
		return
	}

	b, err := yaml.Marshal(&struct {
		Pipeline *ScalingFunctions.PipelineIR `yaml:"pipeline"`
	}{pipeline})
	if err != nil {
		fmt.Printf("Failed to serialize the pipeline '%s': %s\n", identifier, err)
		return
	}

	f, err := os.CreateTemp("", "dfunc-pipeline-*.yml")
	if err != nil {
		fmt.Printf("Failed to create a temporary file for the pipeline '%s': %s\n", identifier, err)
		return
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			fmt.Println(err)
		}
	}(f)

	if _, err := f.Write(b); err != nil {
		fmt.Printf("Failed to write the pipeline '%s' to %s: %s\n", identifier, f.Name(), err)
		return
	}

	// open the pipeline definition in the operator's preferred editor
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to open the pipeline '%s' in %s: %s\n", identifier, editor, err)
	}
}
