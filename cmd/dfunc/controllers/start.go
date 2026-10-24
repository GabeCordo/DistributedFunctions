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

	commandline "github.com/GabeCordo/Commandline"
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
