package controllers

import (
	"fmt"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
)

type MountController struct {
}

func (controller MountController) Run(cli *commandline.CommandLine) (action commandline.TerminateOnCompletion) {

	target := cli.NextArg()
	if target == commandline.FinalArg {
		return commandline.Terminate
	}

	switch target {
	case "function":
		{
			action = controller.mountFunction(cli)
		}
	case "module":
		{
			action = controller.mountModule(cli)
		}
	default:
		{
			action = commandline.Terminate
		}
	}

	return action
}

func (controller MountController) mountFunction(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	module := cli.NextArg()
	if module == commandline.FinalArg {
		fmt.Println("missing module parameter")
		return commandline.Terminate
	}

	function := cli.NextArg()
	if function == commandline.FinalArg {
		fmt.Println("missing function parameter")
		return commandline.Terminate
	}

	if err := api.MountFunction(module, function); err != nil {
		fmt.Printf("Failed to mount the function '%s': %s\n", function, err)
		return commandline.Terminate
	}

	fmt.Printf("The function '%s' has been mounted on the dfunc machine\n", function)

	return commandline.Terminate
}

func (controller MountController) mountModule(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	module := cli.NextArg()
	if module == commandline.FinalArg {
		fmt.Println("missing module parameter")
		return commandline.Terminate
	}

	if err := api.MountModule(module); err != nil {
		fmt.Printf("Failed to mount the module '%s': %s\n", module, err)
		return commandline.Terminate
	}

	fmt.Printf("The module '%s' has been mounted on the dfunc machine\n", module)

	return commandline.Terminate
}
