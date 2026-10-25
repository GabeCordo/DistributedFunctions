package controllers

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"

	commandline "github.com/GabeCordo/Commandline"
)

type PipelinesController struct {
}

func (controller PipelinesController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	return controller.showPipelines(cli)
}

func (controller PipelinesController) showPipelines(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	pipelines, err := api.GetPipelines()
	if err != nil {
		fmt.Printf("Failed to retrieve the pipelines from the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "NAMESPACE\tIDENTIFIER")
	for _, p := range pipelines {
		fmt.Fprintf(w, "%s\t%s\n", p.Namespace, p.Identifier)
	}

	if err := w.Flush(); err != nil {
		fmt.Printf("Failed to print the pipelines: %s\n", err)
	}

	return commandline.Terminate
}
