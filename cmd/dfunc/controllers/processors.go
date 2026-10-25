package controllers

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
)

type ProcessorController struct {
}

func (controller ProcessorController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	processors, err := api.GetProcessors()
	if err != nil {
		fmt.Printf("Failed to retrieve the processors from the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "ID\tADDRESS\tSTATUS\tMODULES\tRUNS")
	for _, p := range processors {
		modules := strings.Join(p.Modules, ", ")
		if modules == "" {
			modules = "-"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\n", p.Id, p.RemoteAddr, p.Status, modules, p.NumOfRuns)
	}

	if err := w.Flush(); err != nil {
		fmt.Printf("Failed to print the processors: %s\n", err)
	}

	return commandline.Terminate
}
