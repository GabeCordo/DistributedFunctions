// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  mount.go
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
		fmt.Println("Missing 'type' parameter: dfunc mount [type(module|function)]")
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
			fmt.Printf("Invalid 'type' Parameter: must be 'module' or 'function' but was '%s'\n", target)
			action = commandline.Terminate
		}
	}

	return action
}

func (controller MountController) mountFunction(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	module := cli.NextArg()
	if module == commandline.FinalArg {
		fmt.Println("Missing 'module' Parameter: dfunc mount function [module] [function]")
		return commandline.Terminate
	}

	function := cli.NextArg()
	if function == commandline.FinalArg {
		fmt.Printf("Missing 'function' Parameter: dfunc mount function %s [function]\n", module)
		return commandline.Terminate
	}

	if err := api.MountFunction(module, function); err != nil {
		fmt.Printf("Failed to mount '%s:%s': %s\n", module, function, err)
		return commandline.Terminate
	}

	fmt.Printf("Succesfully mounted '%s:%s': the function may be run with 'dfunc run'\n", module, function)

	return commandline.Terminate
}

func (controller MountController) mountModule(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	module := cli.NextArg()
	if module == commandline.FinalArg {
		fmt.Println("Missing 'module' Parameter: dfunc mount module [module]")
		return commandline.Terminate
	}

	if err := api.MountModule(module); err != nil {
		fmt.Printf("Failed to mount '%s': %s\n", module, err)
		return commandline.Terminate
	}

	fmt.Printf("Succesfully mounted '%s': the module may be run with 'dfunc run'\n", module)

	return commandline.Terminate
}
