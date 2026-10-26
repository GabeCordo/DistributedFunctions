// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  functions.go
package controllers

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
)

type FunctionController struct {
}

func (controller FunctionController) Run(cli *commandline.CommandLine) (action commandline.TerminateOnCompletion) {

	action = controller.showFunctions(cli)
	return action
}

func (controller FunctionController) showFunctions(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	functions, err := api.GetFunctions()
	if err != nil {
		fmt.Printf("Failed to retrieve the functions from the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "NAME\tMODULE\tPARAMS\tRETURNS\tMOUNTED")
	for _, f := range functions {
		params := strings.Join(f.Parameters, ", ")
		if params == "" {
			params = "-"
		}

		returns := strings.Join(f.Returns, ", ")
		if returns == "" {
			returns = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\n", f.Name, f.Module, params, returns, f.Mounted)
	}

	if err := w.Flush(); err != nil {
		fmt.Printf("Failed to print the functions: %s\n", err)
	}

	return commandline.Terminate
}
