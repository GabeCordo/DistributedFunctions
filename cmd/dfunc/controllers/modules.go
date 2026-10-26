// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  modules.go
package controllers

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
)

type ModuleController struct {
}

func (controller ModuleController) Run(cli *commandline.CommandLine) (action commandline.TerminateOnCompletion) {

	action = controller.getModules(cli)
	return action
}

func (controller ModuleController) getModules(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	modules, err := api.GetModules()
	if err != nil {
		fmt.Printf("Failed to retrieve the modules from the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "NAME\tVERSION\tCONTACT\tMOUNTED")
	for _, m := range modules {
		contact := strings.TrimSpace(m.Contact.Name + " " + m.Contact.Email)
		if contact == "" {
			contact = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%t\n", m.Name, m.Version, contact, m.Mounted)
	}

	if err := w.Flush(); err != nil {
		fmt.Printf("Failed to print the modules: %s\n", err)
	}

	return commandline.Terminate
}
