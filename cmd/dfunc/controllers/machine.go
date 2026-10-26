// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  machine.go
package controllers

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core"
)

type StartCommand struct {
}

const DatabaseUriEnv = "DATABASE_URI"

type EnvironmentVariables struct {
	DatabaseUri string
}

func (sc StartCommand) readEnvironmentVariables() (env EnvironmentVariables) {

	env.DatabaseUri = os.Getenv(DatabaseUriEnv)
	return env
}

func (sc StartCommand) verifyMandatoryEnvironmentVariables(env EnvironmentVariables) (err error) {

	// there are no mandatory environment variables for now.
	err = nil
	return err
}

func (sc StartCommand) Run(cli *commandline.CommandLine) (action commandline.TerminateOnCompletion) {

	runDirective := cli.NextArg()
	blockingFlag := cli.NextArg() == "--blocking"

	switch runDirective {
	case "start":
		{
			if blockingFlag {
				action = sc.blockingStartRun(cli)
			} else {
				action = sc.startRun(cli)
			}
		}
	case "stop":
		{
			action = sc.stopRun(cli)
		}
	case "status":
		{
			action = sc.statusRun(cli)
		}
	default:
		{
			action = commandline.Terminate
		}
	}

	return action
}

func (sc StartCommand) startRun(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	cmd := exec.Command("dfunc", "machine", "start", "--blocking")

	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to start dfunc machine on the host: ", err)
	}

	return commandline.Terminate
}

func (sc StartCommand) statusRun(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	pid, err := api.Ping()
	if err != nil {
		fmt.Println("The dfunc machine is offline. Run 'dfunc machine start' to bring the dfunc machine online")
		return commandline.Terminate
	}

	fmt.Printf("The dfunc machine is running on the host with the PID '%d'\n", pid)

	return commandline.Terminate
}

func (sc StartCommand) stopRun(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	pid, err := api.Ping()
	if err != nil {
		log.Println("Could not find a running instace of dfunc machine on the host")
		return commandline.Terminate
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("Finding the dfunc machine on the host failed with: %s\n", err)
		return commandline.Terminate
	}

	if err := proc.Kill(); err != nil {
		log.Printf("Killing the dfunc machine on the host failed with: %s\n", err)
	}

	return commandline.Terminate
}

func (sc StartCommand) blockingStartRun(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("Failed to start the dfunc machine: the dfunc configurations folder %s is missing\n", DefaultConfigsFolder)
		fmt.Println("Run 'dfunc init' to initialize the dfunc machine.")
		return commandline.Terminate
	}

	fmt.Println(cli.Graphics.Banner)

	env := sc.readEnvironmentVariables()

	err := sc.verifyMandatoryEnvironmentVariables(env)
	if err != nil {
		fmt.Printf("Failed to start the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	cfg, err := core.GetConfigInstance(DefaultCoreConfigFile)
	if err != nil {
		fmt.Printf("Failed to start the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	// Override the database URL if the environment variable is set.
	if env.DatabaseUri != "" {
		log.Printf("The environment variable '%s' will override the config\n", DatabaseUriEnv)
		cfg.Database.Url = env.DatabaseUri
	}

	c, err := core.New(cfg)
	if err != nil {
		log.Printf("Failed to start the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	c.Run()

	return commandline.Terminate
}
