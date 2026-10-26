// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  unmount.go
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
		fmt.Println("Missing 'type' parameter: dfunc unmount [type(module|function)]")
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
			fmt.Printf("Invalid 'type' Parameter: must be 'module' or 'function' but was '%s'\n", target)
			action = commandline.Terminate
		}
	}

	return action
}

func (controller UnMountController) unmountFunction(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	module := cli.NextArg()
	if module == commandline.FinalArg {
		fmt.Println("Missing 'module' Parameter: dfunc unmount function [module] [function]")
		return commandline.Terminate
	}

	function := cli.NextArg()
	if function == commandline.FinalArg {
		fmt.Printf("Missing 'function' Parameter: dfunc unmount function %s [function]\n", module)
		return commandline.Terminate
	}

	if err := api.UnmountFunction(module, function); err != nil {
		fmt.Printf("Failed to unmount '%s:%s': %s\n", module, function, err)
		return commandline.Terminate
	}

	fmt.Printf("Succesfully unmounted '%s:%s'\n", module, function)

	return commandline.Terminate
}

func (controller UnMountController) unmountModule(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	module := cli.NextArg()
	if module == commandline.FinalArg {
		fmt.Println("Missing 'module' Parameter: dfunc unmount module [module]")
		return commandline.Terminate
	}

	if err := api.UnmountModule(module); err != nil {
		fmt.Printf("Failed to unmount '%s': %s\n", module, err)
		return commandline.Terminate
	}

	fmt.Printf("Succesfully unmounted '%s'\n", module)

	return commandline.Terminate
}
