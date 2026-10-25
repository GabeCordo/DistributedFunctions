package controllers

import (
	"fmt"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
)

type UnMountController struct {
}

func (controller UnMountController) Run(cli *commandline.CommandLine) (action commandline.TerminateOnCompletion) {

	target := cli.NextArg()
	if target == commandline.FinalArg {
		return commandline.Terminate
	}

	switch target {
	case "function":
		{
			action = controller.unmountFunction(cli)
		}
	case "module":
		{
			action = controller.unmountModule(cli)
		}
	default:
		{
			action = commandline.Terminate
		}
	}

	return action
}

func (controller UnMountController) unmountFunction(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

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

	if err := api.UnmountFunction(module, function); err != nil {
		fmt.Printf("Failed to unmount the function '%s': %s\n", function, err)
		return commandline.Terminate
	}

	fmt.Printf("The function '%s' has been unmounted on the dfunc machine\n", function)

	return commandline.Terminate
}

func (controller UnMountController) unmountModule(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	module := cli.NextArg()
	if module == commandline.FinalArg {
		fmt.Println("missing module parameter")
		return commandline.Terminate
	}

	if err := api.UnmountModule(module); err != nil {
		fmt.Printf("Failed to unmount the module '%s': %s\n", module, err)
		return commandline.Terminate
	}

	fmt.Printf("The module '%s' has been unmounted on the dfunc machine\n", module)

	return commandline.Terminate
}
