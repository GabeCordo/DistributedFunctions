// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  ps.go
package controllers

import (
	"fmt"
	"os"
	"text/tabwriter"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
)

type PSController struct {
}

func (controller PSController) Run(cli *commandline.CommandLine) (action commandline.TerminateOnCompletion) {

	action = controller.showRuns(cli)
	return action
}

func (controller PSController) showRuns(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	showAllFlag := cli.NextArg() == "-a"

	runs, err := api.GetRuns()
	if err != nil {
		fmt.Printf("Failed to retrieve the runs from the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "ID\tNAMESPACE\tSTATUS\tSTARTED BY\tPROCESSOR\tCREATED")
	for _, r := range runs {
		if showAllFlag || (!showAllFlag && r.IsRunning()) {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\t%s\n",
				r.Id, r.Namespace, r.Status, r.StartedBy, r.Processor, r.Time.Created.Format("2006-01-02 15:04:05"))
		}
	}
	if err := w.Flush(); err != nil {
		fmt.Printf("Failed to print the runs: %s\n", err)
	}

	return commandline.Terminate
}
