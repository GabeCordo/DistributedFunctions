// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  start.go
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
	MongoDbUri string
}

func (sc StartCommand) readEnvironmentVariables() (env EnvironmentVariables) {

	env.MongoDbUri = os.Getenv(DatabaseUriEnv)
	return env
}

func (sc StartCommand) verifyMandatoryEnvironmentVariables(env EnvironmentVariables) (err error) {

	// there are no mandatory environment variables for now.
	err = nil
	return err
}

func (sc StartCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	runDirective := cli.NextArg()
	blockingFlag := cli.NextArg() == "--blocking"

	switch runDirective {
	case "start":
		{
			if blockingFlag {
				return sc.blockingStartRun(cli)
			} else {
				return sc.startRun(cli)
			}
		}
	case "stop":
		{
			return sc.stopRun(cli)
		}
	case "status":
		{
			return sc.statusRun(cli)
		}
	default:
		{
			return commandline.Terminate
		}
	}
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

	// check to see that the etl thread has been initialized with the required files
	// if it has not, fail and tell the operator to call the 'etl init' command
	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("missing configurations folder at %s\nmake sure you run 'DistributedFunctions init'\n", DefaultConfigsFolder)
		return commandline.Terminate
	}

	fmt.Println(cli.Graphics.Banner)

	log.Println("reading environment variables")
	env := sc.readEnvironmentVariables()

	log.Println("verifying environment variables")
	err := sc.verifyMandatoryEnvironmentVariables(env)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}

	log.Printf("attempting to load config instance at %s\n", DefaultCoreConfigFile)
	cfg, err := core.GetConfigInstance(DefaultCoreConfigFile)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}

	// Override the database URL if the environment variable is set.
	if env.MongoDbUri != "" {
		log.Printf("found the environment variable '%s' defined, will override the config\n", DatabaseUriEnv)
		cfg.Database.Url = env.MongoDbUri
	}

	c, err := core.New(cfg)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}

	log.Println("configuration phase is now complete, starting the core")
	log.Println()

	c.Run()

	return commandline.Terminate
}
